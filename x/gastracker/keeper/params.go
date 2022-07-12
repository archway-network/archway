package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	gstTypes "github.com/archway-network/archway/x/gastracker"
)

func (k Keeper) SetParams(ctx sdk.Context, params gstTypes.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

func (k Keeper) IsGasTrackingEnabled(ctx sdk.Context) (res bool) {
	k.paramSpace.Get(ctx, gstTypes.ParamsKeyGasTrackingSwitch, &res)
	return
}

func (k Keeper) IsDappInflationRewardsEnabled(ctx sdk.Context) (res bool) {
	k.paramSpace.Get(ctx, gstTypes.ParamsKeyDappInflationRewards, &res)
	return
}
func (k Keeper) IsGasRebateToContractEnabled(ctx sdk.Context) (res bool) {
	k.paramSpace.Get(ctx, gstTypes.ParamsKeyGasRebateSwitch, &res)
	return
}
func (k Keeper) IsGasRebateToUserEnabled(ctx sdk.Context) (res bool) {
	k.paramSpace.Get(ctx, gstTypes.ParamsKeyGasRebateToUserSwitch, &res)
	return
}
func (k Keeper) IsContractPremiumEnabled(ctx sdk.Context) (res bool) {
	k.paramSpace.Get(ctx, gstTypes.ParamsKeyContractPremiumSwitch, &res)
	return
}
func (k Keeper) InflationRewardQuotaPercentage(ctx sdk.Context) (res uint64) {
	k.paramSpace.Get(ctx, gstTypes.ParamsKeyInflationRewardQuotaPercentage, &res)
	return
}
func (k Keeper) GasRebatePercentage(ctx sdk.Context) (res uint64) {
	k.paramSpace.Get(ctx, gstTypes.ParamsKeyGasRebatePercentage, &res)
	return
}
