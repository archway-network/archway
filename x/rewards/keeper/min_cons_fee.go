package keeper

import (
	"reflect"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/pkg"
	"github.com/archway-network/archway/x/rewards/types"
)

// UpdateMinConsensusFee calculates and updates the minimum consensus fee if eligible emitting an event.
func (k Keeper) UpdateMinConsensusFee(ctx sdk.Context, inflationRewards sdk.Coin) {
	// Prepare and verify inputs
	if inflationRewards.IsZero() {
		k.Logger(ctx).Info("Minimum consensus fee update skipped: inflation rewards are zero")
		return
	}
	inflationRewardsAmt := sdk.NewDecFromInt(inflationRewards.Amount)

	blockGasLimit := ctx.BlockGasMeter().Limit()
	if strings.Contains(reflect.TypeOf(ctx.BlockGasMeter()).String(), "infiniteGasMeter") { // Because thisss https://github.com/cosmos/cosmos-sdk/pull/9651
		blockGasLimit = 0
	}

	blockGasLimitAsDec := pkg.NewDecFromUint64(blockGasLimit)
	if blockGasLimitAsDec.IsZero() {
		k.Logger(ctx).Info("Minimum consensus fee update skipped: block gas limit is not set")
		return
	}

	txFeeRebateRatio := k.TxFeeRebateRatio(ctx)

	// Calculate
	feeAmt := calculateMinConsensusFeeAmt(inflationRewardsAmt, blockGasLimitAsDec, txFeeRebateRatio)
	if feeAmt.IsZero() || feeAmt.IsNegative() {
		k.Logger(ctx).Info("Minimum consensus fee update skipped: calculated amount is zero or bellow zero")
		return
	}
	feeCoin := sdk.DecCoin{
		Denom:  inflationRewards.Denom,
		Amount: feeAmt,
	}

	// Set and emit event
	k.state.MinConsensusFee(ctx).SetFee(feeCoin)
	k.Logger(ctx).Info("Minimum consensus fee update", "fee", feeCoin)

	types.EmitMinConsensusFeeSetEvent(ctx, feeCoin)
}

// GetMinConsensusFee returns the minimum consensus fee.
// Fee defines the minimum gas unit price for a transaction to be included in a block.
func (k Keeper) GetMinConsensusFee(ctx sdk.Context) (sdk.DecCoin, bool) {
	fee, found := k.state.MinConsensusFee(ctx).GetFee()
	if !found {
		return sdk.DecCoin{}, false
	}

	return fee, true
}

// ComputationalPriceOfGas returns the minimum price of each unit of gas.
func (k Keeper) ComputationalPriceOfGas(ctx sdk.Context) sdk.DecCoin {
	minPoG := k.MinimumPriceOfGas(ctx)
	antiDoSPoG, found := k.GetMinConsensusFee(ctx)
	if !found {
		return minPoG
	}
	if minPoG.Denom != antiDoSPoG.Denom {
		panic("conflict between anti dos denom and min price of gas denom: %s != %s" + minPoG.Denom + antiDoSPoG.Denom)
	}
	return sdk.NewDecCoinFromDec(minPoG.Denom, sdk.MaxDec(minPoG.Amount, antiDoSPoG.Amount))
}

// calculateMinConsensusFee calculates the minimum consensus fee amount using the formula:
//
//	[ -1 * ( BlockRewards / ( GasLimit * (TxFeeRatio - 1) ) ]
//
// A simplified expression is used, original from specs: -1 * ( BlockRewards / ( GasLimit * TxFeeRatio - GasLimit ) )
func calculateMinConsensusFeeAmt(blockRewards, gasLimit, txFeeRatio sdk.Dec) sdk.Dec {
	return blockRewards.Quo(
		gasLimit.Mul(
			txFeeRatio.Sub(sdk.OneDec()),
		),
	).Neg()
}
