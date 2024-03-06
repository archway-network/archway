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
	cdc        codec.Codec
	storeKey   storetypes.StoreKey
	wasmKeeper types.WasmKeeperExpected
	authority  string // this should be the x/gov module account

	Schema collections.Schema

	// Params key: ParamsKeyPrefix | value: Params
	Params collections.Item[types.Params]
	// ErrorsCount key: ErrorsCountKey | value: ErrorsCount
	ErrorsCount collections.Item[int64]
	// ContractErrors key: ContractErrorsKeyPrefix + contractAddress + ErrorId | value: ErrorId
	ContractErrors collections.Map[collections.Pair[string, int64], int64]
	// ContractErrors key: ErrorsKeyPrefix + ErrorId | value: SudoError
	Errors collections.Map[int64, types.SudoError]
	// DeletionBlocks key: DeletionBlocksKeyPrefix + BlockHeight + ErrorId | value: ErrorId
	DeletionBlocks collections.Map[collections.Pair[int64, int64], int64]
}

// NewKeeper creates a new Keeper instance.
func NewKeeper(cdc codec.Codec, storeKey storetypes.StoreKey, wk types.WasmKeeperExpected, authority string) Keeper {
	sb := collections.NewSchemaBuilder(collcompat.NewKVStoreService(storeKey))
	k := Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		wasmKeeper: wk,
		authority:  authority,
		Params: collections.NewItem(
			sb,
			types.ParamsKeyPrefix,
			"params",
			collcompat.ProtoValue[types.Params](cdc),
		),
		ErrorsCount: collections.NewItem(
			sb,
			types.ErrorsCountKey,
			"errorsCount",
			collections.Int64Value,
		),
		ContractErrors: collections.NewMap(
			sb,
			types.ContractErrorsKeyPrefix,
			"contractErrors",
			collections.PairKeyCodec(collections.StringKey, collections.Int64Key),
			collections.Int64Value,
		),
		Errors: collections.NewMap(
			sb,
			types.ErrorsKeyPrefix,
			"errors",
			collections.Int64Key,
			collcompat.ProtoValue[types.SudoError](cdc),
		),
		DeletionBlocks: collections.NewMap(
			sb,
			types.DeletionBlocksKeyPrefix,
			"deletionBlocks",
			collections.PairKeyCodec(collections.Int64Key, collections.Int64Key),
			collections.Int64Value,
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

// GetAuthority returns the x/callback module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// SetWasmKeeper sets the given wasm keeper.
// Only for testing purposes
func (k *Keeper) SetWasmKeeper(wk types.WasmKeeperExpected) {
	k.wasmKeeper = wk
}
