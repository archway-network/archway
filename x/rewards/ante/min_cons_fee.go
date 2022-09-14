package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/archway-network/archway/pkg"
)

// MinConsensusFeeReaderExpected defines the expected interface for the x/rewards keeper.
type MinConsensusFeeReaderExpected interface {
	GetMinConsensusFee(ctx sdk.Context) (sdk.DecCoin, bool)
}

// MinFeeDecorator rejects transaction if its fees are less than minimum fees defined by the x/rewards module.
// Estimation is done using the minimum consensus fee value which is the minimum gas unit price.
// The minimum consensus fee value is defined by block dApp rewards and rewards distribution parameters.
// CONTRACT: Tx must implement FeeTx interface to use MinFeeDecorator.
type MinFeeDecorator struct {
	rewardsKeeper MinConsensusFeeReaderExpected
}

// NewMinFeeDecorator returns a new MinFeeDecorator instance.
func NewMinFeeDecorator(rk MinConsensusFeeReaderExpected) MinFeeDecorator {
	return MinFeeDecorator{
		rewardsKeeper: rk,
	}
}

// AnteHandle implements the ante.AnteDecorator interface.
func (mfd MinFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// Skip fee verification for simulation (--dry-run)
	if simulate {
		return next(ctx, tx, simulate)
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkErrors.Wrap(sdkErrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// Skip the check if min gas unit price is not defined (not yet set or is zero)
	gasUnitPrice, found := mfd.rewardsKeeper.GetMinConsensusFee(ctx)
	if !found || gasUnitPrice.IsZero() {
		return next(ctx, tx, simulate)
	}

	// Estimate the minimum fee expected
	// We use RoundInt here since minimum fee must be GTE calculated amount
	txFees := feeTx.GetFee()

	txGasLimit := pkg.NewDecFromUint64(feeTx.GetGas())
	if txGasLimit.IsZero() {
		return ctx, sdkErrors.Wrap(sdkErrors.ErrInvalidRequest, "tx gas limit is not set")
	}

	minFeeExpected := sdk.Coin{
		Denom:  gasUnitPrice.Denom,
		Amount: gasUnitPrice.Amount.Mul(txGasLimit).RoundInt(),
	}

	// Check (skip if the expected amount is zero)
	if minFeeExpected.Amount.IsZero() || txFees.IsAnyGTE(sdk.Coins{minFeeExpected}) {
		return next(ctx, tx, simulate)
	}

	return ctx, sdkErrors.Wrapf(sdkErrors.ErrInsufficientFee, "tx fee %s is less than min fee: %s", txFees, minFeeExpected)
}
