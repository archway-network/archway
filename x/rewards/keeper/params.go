package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/rewards/types"
)

// InflationRewardsRatio return inflation rewards params ratio.
func (k Keeper) InflationRewardsRatio(ctx sdk.Context) (res sdk.Dec) {
	k.paramStore.Get(ctx, types.InflationRewardsRatioParamKey, &res)
	return
}

// InflationRewardsEnabled return inflation rewards enabled flag.
func (k Keeper) InflationRewardsEnabled(ctx sdk.Context) bool {
	return !k.InflationRewardsRatio(ctx).IsZero()
}

// TxFeeRebateRatio return tx fee rebate rewards params ratio.
func (k Keeper) TxFeeRebateRatio(ctx sdk.Context) (res sdk.Dec) {
	k.paramStore.Get(ctx, types.TxFeeRebateRatioParamKey, &res)
	return
}

// TxFeeRewardsEnabled return tx fee rewards enabled flag.
func (k Keeper) TxFeeRewardsEnabled(ctx sdk.Context) bool {
	return !k.TxFeeRebateRatio(ctx).IsZero()
}

// GetParams return all module parameters.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.InflationRewardsRatio(ctx),
		k.TxFeeRebateRatio(ctx),
	)
}

// SetParams sets all module parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramStore.SetParamSet(ctx, &params)
}
