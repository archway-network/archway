package gastracker

import (
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetParams(ctx sdk.Context, params gstTypes.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

func (k Keeper) GetGasTrackingSwitch(ctx sdk.Context) (res bool) {
	return k.paramSpace.Get(ctx, gstTypes.KeyGasTrackingSwitch)
}
func (k Keeper) GetGasRebateSwitch(ctx sdk.Context) (res bool) {
	return k.paramSpace.Get(ctx, gstTypes.KeyGasGasRebateSwitch)
}
func (k Keeper) GetGasRebateToUserSwitch(ctx sdk.Context) (res bool) {
	return k.paramSpace.Get(ctx, gstTypes.KeyGasGasRebateToUserSwitch)
}
func (k Keeper) GetContractPremiumSwitch(ctx sdk.Context) (res bool) {
	return k.paramSpace.Get(ctx, gstTypes.KeyGasontractPremiumSwitch)
}
