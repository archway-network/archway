package keeper

import (
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/callback/types"
)

// Keeper provides module state operations.
type Keeper struct {
	cdc            codec.Codec
	storeKey       storetypes.StoreKey
	rewardsKeepers types.RewardsKeeperExpected
	wasmKeeper     types.WasmKeeperExpected
}

// NewKeeper creates a new Keeper instance.
func NewKeeper(cdc codec.Codec, storeKey storetypes.StoreKey, rk types.RewardsKeeperExpected, wk types.WasmKeeperExpected) Keeper {
	return Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		rewardsKeepers: rk,
		wasmKeeper:     wk,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}
