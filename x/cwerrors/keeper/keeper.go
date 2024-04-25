package keeper

import (
	"cosmossdk.io/collections"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/internal/collcompat"
	"github.com/archway-network/archway/x/cwerrors/types"
)

// Keeper provides module state operations.
type Keeper struct {
	cdc           codec.Codec
	storeKey      storetypes.StoreKey
	tStoreKey     storetypes.StoreKey
	wasmKeeper    types.WasmKeeperExpected
	bankKeeper    types.BankKeeperExpected
	rewardsKeeper types.RewardsKeeperExpected
	authority     string // this should be the x/gov module account

	Schema collections.Schema

	// Params key: ParamsKeyPrefix | value: Params
	Params collections.Item[types.Params]
	// ErrorID key: ErrorsCountKey | value: ErrorID
	ErrorID collections.Sequence
	// ContractErrors key: ContractErrorsKeyPrefix + contractAddress + ErrorId | value: nil
	ContractErrors collections.Map[collections.Pair[[]byte, uint64], []byte]
	// ContractErrors key: ErrorsKeyPrefix + ErrorId | value: SudoError
	Errors collections.Map[uint64, types.SudoError]
	// DeletionBlocks key: DeletionBlocksKeyPrefix + BlockHeight + ErrorId | value: nil
	DeletionBlocks collections.Map[collections.Pair[int64, uint64], []byte]
	// ContractSubscriptions key: ContractSubscriptionsKeyPrefix + contractAddress | value: deletionHeight
	ContractSubscriptions collections.Map[[]byte, int64]
	// SubscriptionEndBlock key: SubscriptionEndBlockKeyPrefix + BlockHeight + contractAddress | value: nil
	SubscriptionEndBlock collections.Map[collections.Pair[int64, []byte], []byte]
}

// NewKeeper creates a new Keeper instance.
func NewKeeper(cdc codec.Codec, storeKey storetypes.StoreKey, tStoreKey storetypes.StoreKey,
	wk types.WasmKeeperExpected, bk types.BankKeeperExpected, rk types.RewardsKeeperExpected,
	authority string,
) Keeper {
	sb := collections.NewSchemaBuilder(collcompat.NewKVStoreService(storeKey))
	k := Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		tStoreKey:     tStoreKey,
		wasmKeeper:    wk,
		bankKeeper:    bk,
		rewardsKeeper: rk,
		authority:     authority,
		Params: collections.NewItem(
			sb,
			types.ParamsKeyPrefix,
			"params",
			collcompat.ProtoValue[types.Params](cdc),
		),
		ErrorID: collections.NewSequence(
			sb,
			types.ErrorIDKey,
			"errorID",
		),
		ContractErrors: collections.NewMap(
			sb,
			types.ContractErrorsKeyPrefix,
			"contractErrors",
			collections.PairKeyCodec(collections.BytesKey, collections.Uint64Key),
			collections.BytesValue,
		),
		Errors: collections.NewMap(
			sb,
			types.ErrorsKeyPrefix,
			"errors",
			collections.Uint64Key,
			collcompat.ProtoValue[types.SudoError](cdc),
		),
		DeletionBlocks: collections.NewMap(
			sb,
			types.DeletionBlocksKeyPrefix,
			"deletionBlocks",
			collections.PairKeyCodec(collections.Int64Key, collections.Uint64Key),
			collections.BytesValue,
		),
		ContractSubscriptions: collections.NewMap(
			sb,
			types.ContractSubscriptionsKeyPrefix,
			"contractSubscriptions",
			collections.BytesKey,
			collections.Int64Value,
		),
		SubscriptionEndBlock: collections.NewMap(
			sb,
			types.SubscriptionEndBlockKeyPrefix,
			"subscriptionEndBlock",
			collections.PairKeyCodec(collections.Int64Key, collections.BytesKey),
			collections.BytesValue,
		),
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetAuthority returns the x/cwerrors module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// SetWasmKeeper sets the given wasm keeper.
// Only for testing purposes
func (k *Keeper) SetWasmKeeper(wk types.WasmKeeperExpected) {
	k.wasmKeeper = wk
}
