package gastracker

import (
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetParams(ctx sdk.Context, params gstTypes.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

func (k Keeper) IsGasTrackingEnabled(ctx sdk.Context) (res bool) {
	k.paramSpace.Get(ctx, gstTypes.KeyGasTrackingSwitch, &res)
	return
}

func (k Keeper) IsDappInflationRewardsEnabled(ctx sdk.Context) (res bool) {
	k.paramSpace.Get(ctx, gstTypes.KeyDappInflationRewards, &res)
	return
}
func (k Keeper) IsGasRebateEnabled(ctx sdk.Context) (res bool) {
	k.paramSpace.Get(ctx, gstTypes.KeyGasRebateSwitch, &res)
	return
}
func (k Keeper) IsGasRebateToUserEnabled(ctx sdk.Context) (res bool) {
	k.paramSpace.Get(ctx, gstTypes.KeyGasRebateToUserSwitch, &res)
	return
}
func (k Keeper) IsContractPremiumEnabled(ctx sdk.Context) (res bool) {
	k.paramSpace.Get(ctx, gstTypes.KeyContractPremiumSwitch, &res)
	return
}
