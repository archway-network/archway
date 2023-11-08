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
}

// NewKeeper creates a new Keeper instance.
func NewKeeper(cdc codec.Codec, storeKey storetypes.StoreKey) Keeper {
	sb := collections.NewSchemaBuilder(collcompat.NewKVStoreService(storeKey))
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		Params:   collections.NewItem(sb, types.ParamsKeyPrefix, "params", collcompat.ProtoValue[types.Params](cdc)),
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}
