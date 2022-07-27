package keeper

import (
	"github.com/archway-network/archway/x/tracking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GasTrackingEnabled return gas tracking enabled param flag.
func (k Keeper) GasTrackingEnabled(ctx sdk.Context) (res bool) {
	k.paramStore.Get(ctx, types.GasTrackingEnabledParamKey, &res)
	return
}

// GetParams return all module parameters.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.GasTrackingEnabled(ctx),
	)
}

// SetParams sets all module parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramStore.SetParamSet(ctx, &params)
}
