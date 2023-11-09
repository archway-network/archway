package keeper

import (
	"cosmossdk.io/collections"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/internal/collcompat"
	"github.com/archway-network/archway/x/callback/types"
)

// Keeper provides module state operations.
type Keeper struct {
	cdc      codec.Codec
	storeKey storetypes.StoreKey
	// Params key: ParamsKeyPrefix | value: Params
	Params collections.Item[types.Params]
  // Callbacks key: CallbackKeyPrefix | value: []Callback
  Callbacks collections.Map[collections.Triple[int64, []byte, uint64], types.Callback]
}

// NewKeeper creates a new Keeper instance.
func NewKeeper(cdc codec.Codec, storeKey storetypes.StoreKey) Keeper {
	sb := collections.NewSchemaBuilder(collcompat.NewKVStoreService(storeKey))
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		Params:   collections.NewItem(sb, types.ParamsKeyPrefix, "params", collcompat.ProtoValue[types.Params](cdc)),
    Callbacks: collections.NewMap(
			schemaBuilder,
			types.CallbackKeyPrefix,
			"callbacks",
			collections.TripleKeyCodec(collections.Int64Key, collections.BytesKey, collections.Uint64Key),
			collcompat.ProtoValue[types.Callback](cdc),
		),
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}
