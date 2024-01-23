package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/archway-network/archway/x/rewards/types"
)

// WithdrawRewardsByRecordsLimit performs the rewards distribution for the given rewards address and the number of record to use.
func (k Keeper) WithdrawRewardsByRecordsLimit(ctx sdk.Context, rewardsAddr sdk.AccAddress, recordsLimit uint64) (sdk.Coins, int, error) {
	recordsLimitMax := k.MaxWithdrawRecords(ctx)

	// Use the default limit if not specified
	if recordsLimit == 0 {
		recordsLimit = recordsLimitMax
	}

	// Msg post-validateBasic check
	if recordsLimit > recordsLimitMax {
		return nil, 0, errorsmod.Wrapf(types.ErrInvalidRequest, "max withdraw records (%d) exceeded", recordsLimitMax)
	}

	// Get all rewards records for the given address by limit
	pageReq := &query.PageRequest{Limit: recordsLimit}
	records, _, err := k.GetRewardsRecordsByWithdrawAddressPaginated(ctx, rewardsAddr, pageReq)
	if err != nil {
		return nil, 0, errorsmod.Wrap(types.ErrInternal, err.Error())
	}

	return k.withdrawRewardsByRecords(ctx, rewardsAddr, records), len(records), nil
}

// WithdrawRewardsByRecordIDs performs the rewards distribution for the given rewards address and record IDs.
func (k Keeper) WithdrawRewardsByRecordIDs(ctx sdk.Context, rewardsAddr sdk.AccAddress, recordIDs []uint64) (sdk.Coins, int, error) {
	// Msg post-validateBasic check
	if maxRecords := k.MaxWithdrawRecords(ctx); uint64(len(recordIDs)) > maxRecords {
		return nil, 0, errorsmod.Wrapf(types.ErrInvalidRequest, "max withdraw records (%d) exceeded", maxRecords)
	}

	rewardsAddrStr := rewardsAddr.String()

	// Check that provided IDs do exist and belong to the given address
	records := make([]types.RewardsRecord, 0, len(recordIDs))
	for _, id := range recordIDs {
		record, err := k.RewardsRecords.Get(ctx, id)
		if err != nil {
			return nil, 0, errorsmod.Wrapf(types.ErrInvalidRequest, "rewards record (%d): not found", id)
		}
		if record.RewardsAddress != rewardsAddrStr {
			return nil, 0, errorsmod.Wrapf(types.ErrInvalidRequest, "rewards record (%d): address mismatch", id)
		}

		records = append(records, record)
	}

	return k.withdrawRewardsByRecords(ctx, rewardsAddr, records), len(records), nil
}

// withdrawRewardsByRecords performs the rewards distribution for the given rewards address and records.
// Handler emits the distribution event and prunes the used records.
func (k Keeper) withdrawRewardsByRecords(ctx sdk.Context, rewardsAddr sdk.AccAddress, records []types.RewardsRecord) sdk.Coins {
	// Aggregate total rewards to distribute
	totalRewards := sdk.NewCoins()
	for _, record := range records {
		totalRewards = totalRewards.Add(record.Rewards...)
	}

	// Transfer rewards and emit distribution event
	if !totalRewards.IsZero() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ContractRewardCollector, rewardsAddr, totalRewards); err != nil {
			panic(fmt.Errorf("sending rewards (%s) to the rewards address (%s): %w", totalRewards, rewardsAddr, err))
		}

		types.EmitRewardsWithdrawEvent(ctx, rewardsAddr, totalRewards)
	}

	// Clean up (safe if there were no rewards)
	err := fastRemoveRecords(ctx, k.storeKey, k.RewardsRecords, records...)
	if err != nil {
		panic(fmt.Errorf("removing rewards records: %w", err))
	}
	return totalRewards
}

// fastRemoveRecords is used to remove rewards records without going through the indexed map
// which fetches the records from the store. This is used in the case where we know the records.
func fastRemoveRecords(ctx sdk.Context, storeKey storetypes.StoreKey, im *collections.IndexedMap[uint64, types.RewardsRecord, RewardsRecordsIndex], records ...types.RewardsRecord) error {
	store := ctx.KVStore(storeKey)
	primaryKeyCodec := im.KeyCodec()
	secondaryKeyCodec := im.Indexes.Address.KeyCodec()

	for _, record := range records {
		primaryKeyBytes, err := collections.EncodeKeyWithPrefix(types.RewardsRecordStatePrefix, primaryKeyCodec, record.Id)
		if err != nil {
			return err
		}
		rewardAddr, err := sdk.AccAddressFromBech32(record.RewardsAddress)
		if err != nil {
			return err
		}
		secondaryKey := collections.Join([]byte(rewardAddr), record.Id)
		secondaryKeyBytes, err := collections.EncodeKeyWithPrefix(types.RewardsRecordAddressIndexPrefix, secondaryKeyCodec, secondaryKey)
		if err != nil {
			return err
		}

		store.Delete(primaryKeyBytes)
		store.Delete(secondaryKeyBytes)
	}

	return nil
}
