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
	// Errors key: CallbackKeyPrefix | value: []Callback
	//Errors collections.Map[collections.Triple[int64, []byte, uint64], types.Callback]
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
		// Errors: collections.NewMap(
		// 	sb,
		// 	types.CallbackKeyPrefix,
		// 	"callbacks",
		// 	collections.TripleKeyCodec(collections.Int64Key, collections.BytesKey, collections.Uint64Key),
		// 	collcompat.ProtoValue[types.Callback](cdc),
		// ),
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
