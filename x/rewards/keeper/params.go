package keeper

import (
	"github.com/archway-network/archway/x/rewards/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RewardsEnabled return rewards calculation and distribution enabled param flag.
func (k Keeper) RewardsEnabled(ctx sdk.Context) (res bool) {
	k.paramStore.Get(ctx, types.RewardsEnabledParamKey, &res)
	return
}

// InflationRewardsRatio return inflation rewards params ratio.
func (k Keeper) InflationRewardsRatio(ctx sdk.Context) (res sdk.Dec) {
	k.paramStore.Get(ctx, types.InflationRewardsRatioParamKey, &res)
	return
}

// TxFeeRebateRatio return tx fee rebate rewards params ratio.
func (k Keeper) TxFeeRebateRatio(ctx sdk.Context) (res sdk.Dec) {
	k.paramStore.Get(ctx, types.TxFeeRebateRatioParamKey, &res)
	return
}

// GetParams return all module parameters.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.RewardsEnabled(ctx),
		k.InflationRewardsRatio(ctx),
		k.TxFeeRebateRatio(ctx),
	)
}

// SetParams sets all module parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramStore.SetParamSet(ctx, &params)
}
