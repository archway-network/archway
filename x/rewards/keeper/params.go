package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/rewards/types"
)

// InflationRewardsRatio return inflation rewards params ratio.
func (k Keeper) InflationRewardsRatio(ctx sdk.Context) (res sdk.Dec) {
	return k.GetParams(ctx).InflationRewardsRatio
}

// TxFeeRebateRatio return tx fee rebate rewards params ratio.
func (k Keeper) TxFeeRebateRatio(ctx sdk.Context) (res sdk.Dec) {
	return k.GetParams(ctx).TxFeeRebateRatio
}

// MaxWithdrawRecords return the maximum number of types.RewardsRecord objects used for the withdrawal operation.
func (k Keeper) MaxWithdrawRecords(ctx sdk.Context) (res uint64) {
	return k.GetParams(ctx).MaxWithdrawRecords
}

func (k Keeper) MinimumPriceOfGas(ctx sdk.Context) sdk.DecCoin {
	return k.GetParams(ctx).MinPriceOfGas
}

// GetParams return all module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	params, _ = k.Params.Get(ctx)
	return
}

// SetParams sets all module parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	store := ctx.KVStore(k.state.key)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey, bz)
	return nil
}
