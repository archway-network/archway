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

// TxFeeRebateRatio return tx fee rebate rewards params ratio.
func (k Keeper) TxFeeRebateRatio(ctx sdk.Context) (res sdk.Dec) {
	k.paramStore.Get(ctx, types.TxFeeRebateRatioParamKey, &res)
	return
}

// MaxWithdrawRecords return the maximum number of types.RewardsRecord objects used for the withdrawal operation.
func (k Keeper) MaxWithdrawRecords(ctx sdk.Context) (res uint64) {
	k.paramStore.Get(ctx, types.MaxWithdrawRecordsParamKey, &res)
	return
}

// GetParams return all module parameters.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.InflationRewardsRatio(ctx),
		k.TxFeeRebateRatio(ctx),
		k.MaxWithdrawRecords(ctx),
	)
}

// SetParams sets all module parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramStore.SetParamSet(ctx, &params)
}
