package keeper

import (
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
	inflationRewardsAmt := inflationRewards.Amount.ToDec()

	blockGasLimit := pkg.NewDecFromUint64(ctx.BlockGasMeter().Limit())
	if blockGasLimit.IsZero() {
		k.Logger(ctx).Info("Minimum consensus fee update skipped: block gas limit is not set")
		return
	}

	txFeeRebateRatio := k.TxFeeRebateRatio(ctx)

	// Calculate
	feeAmt := calculateMinConsensusFeeAmt(inflationRewardsAmt, blockGasLimit, txFeeRebateRatio)
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

// calculateMinConsensusFee calculates the minimum consensus fee amount using the formula:
//
//	-1 * ( BlockRewards / ( GasLimit * (TxFeeRatio - 1) ) )
//
// A simplified expression is used, original from specs: -1 * ( BlockRewards / ( GasLimit * TxFeeRatio - GasLimit ) )
func calculateMinConsensusFeeAmt(blockRewards, gasLimit, txFeeRatio sdk.Dec) sdk.Dec {
	return blockRewards.Quo(
		gasLimit.Mul(
			txFeeRatio.Sub(sdk.OneDec()),
		),
	).Neg()
}
