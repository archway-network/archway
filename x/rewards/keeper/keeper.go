package keeper

import (
	"context"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/query"
	paramTypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/archway-network/archway/internal/collcompat"

	"github.com/archway-network/archway/x/rewards/types"
	trackingTypes "github.com/archway-network/archway/x/tracking/types"
)

// ContractInfoReaderExpected defines the interface for the x/wasmd module dependency.
type ContractInfoReaderExpected interface {
	GetContractInfo(ctx context.Context, contractAddress sdk.AccAddress) *wasmTypes.ContractInfo
}

// TrackingKeeperExpected defines the interface for the x/tracking module dependency.
type TrackingKeeperExpected interface {
	GetCurrentTxID(ctx sdk.Context) uint64
	GetBlockTrackingInfo(ctx sdk.Context, height int64) trackingTypes.BlockTracking
	RemoveBlockTrackingInfo(ctx sdk.Context, height int64)
}

// AuthKeeperExpected defines the interface for the x/auth module dependency.
type AuthKeeperExpected interface {
	GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
}

// BankKeeperExpected defines the interface for the x/bank module dependency.
type BankKeeperExpected interface {
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
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

type RewardsRecordsIndex struct {
	// Address maps the rewards record to the address of the recipient.
	Address *indexes.Multi[[]byte, uint64, types.RewardsRecord]
}

func (t RewardsRecordsIndex) IndexesList() []collections.Index[uint64, types.RewardsRecord] {
	return []collections.Index[uint64, types.RewardsRecord]{t.Address}
}

func NewRewardsRecordsIndex(sb *collections.SchemaBuilder) RewardsRecordsIndex {
	return RewardsRecordsIndex{
		Address: indexes.NewMulti(sb, types.RewardsRecordAddressIndexPrefix, "rewards_records_by_address", collections.BytesKey, collections.Uint64Key, func(_ uint64, value types.RewardsRecord) ([]byte, error) {
			return sdk.AccAddressFromBech32(value.RewardsAddress)
		}),
	}
}

// Keeper provides module state operations.
type Keeper struct {
	cdc              codec.Codec
	storeKey         storetypes.StoreKey
	paramStore       paramTypes.Subspace
	contractInfoView ContractInfoReaderExpected
	trackingKeeper   TrackingKeeperExpected
	authKeeper       AuthKeeperExpected
	bankKeeper       BankKeeperExpected
	authority        string // this should be the x/gov module account
	logger           log.Logger

	Schema collections.Schema

	Params           collections.Item[types.Params]
	MinConsFee       collections.Item[sdk.DecCoin]
	ContractMetadata collections.Map[[]byte, types.ContractMetadata]
	BlockRewards     collections.Map[uint64, types.BlockRewards]
	FlatFees         collections.Map[[]byte, sdk.Coin]
	TxRewards        *collections.IndexedMap[uint64, types.TxRewards, TxRewardsIndex]
	RewardsRecordID  collections.Sequence
	RewardsRecords   *collections.IndexedMap[uint64, types.RewardsRecord, RewardsRecordsIndex]
}

// NewKeeper creates a new Keeper instance.
func NewKeeper(
	cdc codec.Codec,
	key storetypes.StoreKey,
	contractInfoReader ContractInfoReaderExpected,
	trackingKeeper TrackingKeeperExpected,
	ak AuthKeeperExpected,
	bk BankKeeperExpected,
	ps paramTypes.Subspace,
	authority string,
	logger log.Logger,
) Keeper {
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	schemaBuilder := collections.NewSchemaBuilder(collcompat.NewKVStoreService(key))

	k := Keeper{
		storeKey:         key,
		cdc:              cdc,
		paramStore:       ps,
		contractInfoView: contractInfoReader,
		trackingKeeper:   trackingKeeper,
		authKeeper:       ak,
		bankKeeper:       bk,
		authority:        authority,
		logger:           logger.With("module", "x/"+types.ModuleName),
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
		RewardsRecordID: collections.NewSequence(schemaBuilder, types.RewardsRecordsIDPrefix, "rewards_record_id"),
		RewardsRecords: collections.NewIndexedMap(
			schemaBuilder,
			types.RewardsRecordStatePrefix,
			"rewards_records",
			collections.Uint64Key,
			collcompat.ProtoValue[types.RewardsRecord](cdc),
			NewRewardsRecordsIndex(schemaBuilder),
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
	return k.logger
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

	return k.GetRewardsRecordsByWithdrawAddressPaginated(ctx, rewardsAddr, pageReq)
}

func (k Keeper) GetTxRewardsByBlock(ctx context.Context, height uint64) ([]types.TxRewards, error) {
	iter, err := k.TxRewards.Indexes.Block.MatchExact(ctx, height)
	if err != nil {
		return nil, err
	}
	return indexes.CollectValues(ctx, k.TxRewards, iter)
}

// CreateRewardsRecord creates a new rewards record.
func (k Keeper) CreateRewardsRecord(
	ctx context.Context,
	withdrawAddr sdk.AccAddress,
	rewards sdk.Coins,
	calculatedHeight int64,
	calculatedTime time.Time,
) (types.RewardsRecord, error) {
	nextRewardsID, err := k.RewardsRecordID.Next(ctx)
	if err != nil {
		return types.RewardsRecord{}, err
	}
	obj := types.RewardsRecord{
		Id:               nextRewardsID + 1, // rewards record id starts from 1, collections.Sequence starts from 0
		RewardsAddress:   withdrawAddr.String(),
		Rewards:          rewards,
		CalculatedHeight: calculatedHeight,
		CalculatedTime:   calculatedTime,
	}
	return obj, k.RewardsRecords.Set(ctx, obj.Id, obj)
}

// GetRewardsRecordsByWithdrawAddress returns all the rewards records for a given withdraw address.
func (k Keeper) GetRewardsRecordsByWithdrawAddress(ctx context.Context, address sdk.AccAddress) ([]types.RewardsRecord, error) {
	iter, err := k.RewardsRecords.Indexes.Address.MatchExact(ctx, address)
	if err != nil {
		return nil, err
	}
	return indexes.CollectValues(ctx, k.RewardsRecords, iter)
}

// GetRewardsRecordsByWithdrawAddressPaginated returns all the rewards records for a given withdraw address paginated.
// TODO: on v050 replace this with collections paginated.
func (k Keeper) GetRewardsRecordsByWithdrawAddressPaginated(ctx sdk.Context, addr sdk.AccAddress, pageReq *query.PageRequest) ([]types.RewardsRecord, *query.PageResponse, error) {
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		append(types.RewardsRecordAddressIndexPrefix, address.MustLengthPrefix(addr)...),
	)
	var objs []types.RewardsRecord
	pageRes, err := query.Paginate(store, pageReq, func(key, _ []byte) error {
		obj, err := k.RewardsRecords.Get(ctx, sdk.BigEndianToUint64(key))
		if err != nil {
			return err
		}
		objs = append(objs, obj)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return objs, pageRes, nil
}

// GetAuthority returns the x/rewards module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}
