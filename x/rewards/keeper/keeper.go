package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	errorsmod "cosmossdk.io/errors"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/archway-network/archway/internal/collcompat"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramTypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/archway-network/archway/x/rewards/types"
	trackingTypes "github.com/archway-network/archway/x/tracking/types"
)

// ContractInfoReaderExpected defines the interface for the x/wasmd module dependency.
type ContractInfoReaderExpected interface {
	GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmTypes.ContractInfo
}

// TrackingKeeperExpected defines the interface for the x/tracking module dependency.
type TrackingKeeperExpected interface {
	GetCurrentTxID(ctx sdk.Context) uint64
	GetBlockTrackingInfo(ctx sdk.Context, height int64) trackingTypes.BlockTracking
	RemoveBlockTrackingInfo(ctx sdk.Context, height int64)
}

// AuthKeeperExpected defines the interface for the x/auth module dependency.
type AuthKeeperExpected interface {
	GetModuleAccount(ctx sdk.Context, name string) authTypes.ModuleAccountI
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authTypes.AccountI
}

// BankKeeperExpected defines the interface for the x/bank module dependency.
type BankKeeperExpected interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
}

func NewTxRewardsIndex(sb *collections.SchemaBuilder) TxRewardsIndex {
	return TxRewardsIndex{
		Block: indexes.NewMulti(sb, types.TxRewardsHeightIndexPrefix, "tx_rewards_by_block", collections.Uint64Key, collections.Uint64Key, func(_ uint64, value types.TxRewards) (uint64, error) {
			return uint64(value.Height), nil
		}),
	}
}

type TxRewardsIndex struct {
	// Block is the index that maps block height to the TxRewards for that block.
	Block *indexes.Multi[uint64, uint64, types.TxRewards]
}

func (t TxRewardsIndex) IndexesList() []collections.Index[uint64, types.TxRewards] {
	return []collections.Index[uint64, types.TxRewards]{t.Block}
}

// Keeper provides module state operations.
type Keeper struct {
	cdc              codec.Codec
	paramStore       paramTypes.Subspace
	state            State
	contractInfoView ContractInfoReaderExpected
	trackingKeeper   TrackingKeeperExpected
	authKeeper       AuthKeeperExpected
	bankKeeper       BankKeeperExpected
	authority        string // this should be the x/gov module account

	Schema collections.Schema

	Params           collections.Item[types.Params]
	MinConsFee       collections.Item[sdk.DecCoin]
	ContractMetadata collections.Map[[]byte, types.ContractMetadata]
	BlockRewards     collections.Map[uint64, types.BlockRewards]
	FlatFees         collections.Map[[]byte, sdk.Coin]
	TxRewards        *collections.IndexedMap[uint64, types.TxRewards, TxRewardsIndex]
}

// NewKeeper creates a new Keeper instance.
func NewKeeper(cdc codec.Codec, key storetypes.StoreKey, contractInfoReader ContractInfoReaderExpected, trackingKeeper TrackingKeeperExpected, ak AuthKeeperExpected, bk BankKeeperExpected, ps paramTypes.Subspace, authority string) Keeper {
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	schemaBuilder := collections.NewSchemaBuilder(collcompat.NewKVStoreService(key))

	k := Keeper{
		cdc:              cdc,
		paramStore:       ps,
		state:            NewState(cdc, key),
		contractInfoView: contractInfoReader,
		trackingKeeper:   trackingKeeper,
		authKeeper:       ak,
		bankKeeper:       bk,
		authority:        authority,
		Params: collections.NewItem(
			schemaBuilder,
			types.ParamsPrefix,
			"params",
			collcompat.ProtoValue[types.Params](cdc),
		),
		MinConsFee: collections.NewItem(
			schemaBuilder,
			types.MinConsFeePrefix,
			"min_consensus_fee",
			collcompat.ProtoValue[sdk.DecCoin](cdc),
		),
		BlockRewards: collections.NewMap(
			schemaBuilder,
			types.BlockRewardsPrefix,
			"block_rewards",
			collections.Uint64Key,
			collcompat.ProtoValue[types.BlockRewards](cdc),
		),
		ContractMetadata: collections.NewMap(
			schemaBuilder,
			types.ContractMetadataPrefix,
			"contract_metadata",
			collections.BytesKey,
			collcompat.ProtoValue[types.ContractMetadata](cdc),
		),
		FlatFees: collections.NewMap(
			schemaBuilder,
			types.FlatFeePrefix,
			"flat_fees",
			collections.BytesKey,
			collcompat.ProtoValue[sdk.Coin](cdc),
		),
		TxRewards: collections.NewIndexedMap(
			schemaBuilder,
			types.TxRewardsPrefix,
			"tx_rewards",
			collections.Uint64Key,
			collcompat.ProtoValue[types.TxRewards](cdc),
			NewTxRewardsIndex(schemaBuilder),
		),
	}

	schema, err := schemaBuilder.Build()
	if err != nil {
		panic(err)
	}

	k.Schema = schema
	return k
}

// SetContractInfoViewer sets the contract info view dependency.
// Only for testing purposes.
func (k *Keeper) SetContractInfoViewer(viewer ContractInfoReaderExpected) {
	k.contractInfoView = viewer
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// UndistributedRewardsPool returns the current undistributed rewards (yet to be withdrawn).
func (k Keeper) UndistributedRewardsPool(ctx sdk.Context) sdk.Coins {
	poolAcc := k.authKeeper.GetModuleAccount(ctx, types.ContractRewardCollector)
	return k.bankKeeper.GetAllBalances(ctx, poolAcc.GetAddress())
}

// TreasuryPool returns the current undistributed treasury rewards.
func (k Keeper) TreasuryPool(ctx sdk.Context) sdk.Coins {
	poolAcc := k.authKeeper.GetModuleAccount(ctx, types.TreasuryCollector)
	return k.bankKeeper.GetAllBalances(ctx, poolAcc.GetAddress())
}

// GetRewardsRecords returns all the rewards records for a given rewards address paginated.
// Query checks the page limit and uses the default limit if not provided.
func (k Keeper) GetRewardsRecords(ctx sdk.Context, rewardsAddr sdk.AccAddress, pageReq *query.PageRequest) ([]types.RewardsRecord, *query.PageResponse, error) {
	if pageReq == nil {
		pageReq = &query.PageRequest{
			Limit: types.MaxRecordsQueryLimit,
		}
	}
	if pageReq.Limit > types.MaxRecordsQueryLimit {
		return nil, nil, errorsmod.Wrapf(types.ErrInvalidRequest, "max records (%d) query limit exceeded", types.MaxRecordsQueryLimit)
	}

	return k.state.RewardsRecord(ctx).GetRewardsRecordByRewardsAddressPaginated(rewardsAddr, pageReq)
}

func (k Keeper) GetTxRewardsByBlock(ctx context.Context, height uint64) ([]types.TxRewards, error) {
	iter, err := k.TxRewards.Indexes.Block.MatchExact(ctx, height)
	if err != nil {
		return nil, err
	}
	return indexes.CollectValues(ctx, k.TxRewards, iter)
}

// GetAuthority returns the x/rewards module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}
