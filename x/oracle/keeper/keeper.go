package keeper

import (
	"context"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/collections"

	"github.com/archway-network/archway/internal/collcompat"
	archcodec "github.com/archway-network/archway/types/codec"
	"github.com/archway-network/archway/x/oracle/asset"
	"github.com/archway-network/archway/x/oracle/types"
)

// Keeper of the oracle store
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	AccountKeeper  types.AccountKeeper
	bankKeeper     types.BankKeeper
	distrKeeper    types.DistributionKeeper
	StakingKeeper  types.StakingKeeper
	slashingKeeper types.SlashingKeeper

	distrModuleName string

	// Module parameters
	Params            collections.Item[types.Params]
	ExchangeRates     collections.Map[asset.Pair, types.DatedPrice]
	FeederDelegations collections.Map[[]byte, []byte]
	MissCounters      collections.Map[[]byte, uint64]
	Prevotes          collections.Map[[]byte, types.AggregateExchangeRatePrevote]
	Votes             collections.Map[[]byte, types.AggregateExchangeRateVote]

	// PriceSnapshots maps types.PriceSnapshot to the asset.Pair of the snapshot and the creation timestamp as keys.Uint64Key.
	PriceSnapshots collections.Map[
		collections.Pair[asset.Pair, time.Time],
		types.PriceSnapshot]
	WhitelistedPairs collections.KeySet[asset.Pair]
	Rewards          collections.Map[uint64, types.Rewards]
	RewardsID        collections.Sequence
}

// NewKeeper constructs a new keeper for oracle
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,

	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	distrKeeper types.DistributionKeeper,
	stakingKeeper types.StakingKeeper,
	slashingKeeper types.SlashingKeeper,

	distrName string,
) Keeper {
	// ensure oracle module account is set
	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	sb := collections.NewSchemaBuilder(collcompat.NewKVStoreService(storeKey))

	k := Keeper{
		cdc:               cdc,
		storeKey:          storeKey,
		AccountKeeper:     accountKeeper,
		bankKeeper:        bankKeeper,
		distrKeeper:       distrKeeper,
		StakingKeeper:     stakingKeeper,
		slashingKeeper:    slashingKeeper,
		distrModuleName:   distrName,
		Params:            collections.NewItem(sb, collections.NewPrefix(11), "Params", collcompat.ProtoValue[types.Params](cdc)),
		ExchangeRates:     collections.NewMap(sb, collections.NewPrefix(1), "ExchangeRates", asset.PairKeyEncoder, collcompat.ProtoValue[types.DatedPrice](cdc)),
		PriceSnapshots:    collections.NewMap(sb, collections.NewPrefix(10), "PriceSnapshots", collections.PairKeyCodec(asset.PairKeyEncoder, archcodec.TimeKeyEncoder), collcompat.ProtoValue[types.PriceSnapshot](cdc)),
		FeederDelegations: collections.NewMap(sb, collections.NewPrefix(2), "FeederDelegations", collections.BytesKey, collections.BytesValue),
		MissCounters:      collections.NewMap(sb, collections.NewPrefix(3), "MissCounters", collections.BytesKey, collections.Uint64Value),
		Prevotes:          collections.NewMap(sb, collections.NewPrefix(4), "Prevotes", collections.BytesKey, collcompat.ProtoValue[types.AggregateExchangeRatePrevote](cdc)),
		Votes:             collections.NewMap(sb, collections.NewPrefix(5), "Votes", collections.BytesKey, collcompat.ProtoValue[types.AggregateExchangeRateVote](cdc)),
		WhitelistedPairs:  collections.NewKeySet(sb, collections.NewPrefix(6), "WhitelistedPairs", asset.PairKeyEncoder),
		Rewards:           collections.NewMap(sb, collections.NewPrefix(7), "Rewards", collections.Uint64Key, collcompat.ProtoValue[types.Rewards](cdc)),
		RewardsID:         collections.NewSequence(sb, collections.NewPrefix(9), "RewardsID"),
	}
	return k
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ValidateFeeder return the given feeder is allowed to feed the message or not
func (k Keeper) ValidateFeeder(
	ctx context.Context, feederAddr sdk.AccAddress, validatorAddr sdk.ValAddress,
) error {
	// A validator delegates price feeder consent to itself by default.
	// Thus, we only need to verify consent for price feeder addresses that don't
	// match the validator address.
	if !feederAddr.Equals(validatorAddr) {
		delegateStr, err := k.FeederDelegations.Get(
			sdk.UnwrapSDKContext(ctx),
			validatorAddr,
		)
		var delegate sdk.Address
		if err == nil {
			delegate = sdk.AccAddress(delegateStr)
		} else {
			delegate = sdk.AccAddress(validatorAddr)
		}
		if !delegate.Equals(feederAddr) {
			return sdkerrors.Wrapf(
				types.ErrNoVotingPermission,
				"wanted: %s, got: %s", delegate.String(), feederAddr.String())
		}
	}

	// Check that the given validator is in the active set for consensus.
	val, err := k.StakingKeeper.Validator(ctx, validatorAddr)
	if err != nil {
		return sdkerrors.Wrapf(
			err,
			"failed to get validator %s",
			validatorAddr.String(),
		)
	}
	if !val.IsBonded() {
		return sdkerrors.Wrapf(
			fmt.Errorf("Invalid Validator Status"),
			"validator %s is not active set",
			validatorAddr.String(),
		)
	}

	return nil
}

func (k Keeper) GetExchangeRateTwap(ctx sdk.Context, pair asset.Pair) (price math.LegacyDec, err error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return math.LegacyOneDec().Neg(), err
	}

	snapshotsIter, err := k.PriceSnapshots.Iterate(
		ctx,
		collections.NewPrefixedPairRange[asset.Pair, time.Time](pair).
			StartInclusive(
				ctx.BlockTime().Add(-1*params.TwapLookbackWindow)).
			EndInclusive(
				ctx.BlockTime()),
	)
	snapshots, err := snapshotsIter.Values()

	if len(snapshots) == 0 {
		// if there are no snapshots, return -1 for the price
		return math.LegacyOneDec().Neg(), types.ErrNoValidTWAP.Wrapf("no snapshots for pair %s", pair.String())
	}

	if len(snapshots) == 1 {
		return snapshots[0].Price, nil
	}

	firstTimestampMs := snapshots[0].TimestampMs
	if firstTimestampMs > ctx.BlockTime().UnixMilli() {
		// should never happen, or else we have corrupted state
		return math.LegacyOneDec().Neg(), types.ErrNoValidTWAP.Wrapf(
			"Possible corrupted state. First timestamp %d is after current blocktime %d", firstTimestampMs, ctx.BlockTime().UnixMilli())
	}

	if firstTimestampMs == ctx.BlockTime().UnixMilli() {
		// shouldn't happen because we check for len(snapshots) == 1, but if it does, return the first snapshot price
		return snapshots[0].Price, nil
	}

	cumulativePrice := math.LegacyZeroDec()
	for i, s := range snapshots {
		var nextTimestampMs int64
		if i == len(snapshots)-1 {
			// if we're at the last snapshot, then consider that price as ongoing until the current blocktime
			nextTimestampMs = ctx.BlockTime().UnixMilli()
		} else {
			nextTimestampMs = snapshots[i+1].TimestampMs
		}

		price := s.Price.MulInt64(nextTimestampMs - s.TimestampMs)
		cumulativePrice = cumulativePrice.Add(price)
	}

	return cumulativePrice.QuoInt64(ctx.BlockTime().UnixMilli() - firstTimestampMs), nil
}

func (k Keeper) GetExchangeRate(ctx sdk.Context, pair asset.Pair) (price math.LegacyDec, err error) {
	exchangeRate, err := k.ExchangeRates.Get(ctx, pair)
	price = exchangeRate.ExchangeRate
	return
}

// SetPrice sets the price for a pair as well as the price snapshot.
func (k Keeper) SetPrice(ctx sdk.Context, pair asset.Pair, price math.LegacyDec) {
	k.ExchangeRates.Set(ctx, pair, types.DatedPrice{
		ExchangeRate:   price,
		CreationHeight: ctx.BlockHeight(),
		CreationTime:   ctx.BlockTime(),
	})

	key := collections.Join(pair, ctx.BlockTime())
	timestampMs := ctx.BlockTime().UnixMilli()
	k.PriceSnapshots.Set(ctx, key, types.PriceSnapshot{
		Pair:        pair,
		Price:       price,
		TimestampMs: timestampMs,
	})
	if err := ctx.EventManager().EmitTypedEvent(&types.EventPriceUpdate{
		Pair:        pair.String(),
		Price:       price,
		TimestampMs: timestampMs,
	}); err != nil {
		ctx.Logger().Error("failed to emit OraclePriceUpdate", "pair", pair, "error", err)
	}
}
