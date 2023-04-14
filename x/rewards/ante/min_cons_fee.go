package ante

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/archway-network/archway/pkg"
)

// MinFeeDecorator rejects transaction if its fees are less than minimum fees defined by the x/rewards module.
// Estimation is done using the minimum consensus fee value which is the minimum gas unit price.
// The minimum consensus fee value is defined by block dApp rewards and rewards distribution parameters.
// CONTRACT: Tx must implement FeeTx interface to use MinFeeDecorator.
type MinFeeDecorator struct {
	codec         codec.BinaryCodec
	rewardsKeeper RewardsKeeperExpected
}

// NewMinFeeDecorator returns a new MinFeeDecorator instance.
func NewMinFeeDecorator(codec codec.BinaryCodec, rk RewardsKeeperExpected) MinFeeDecorator {
	return MinFeeDecorator{
		codec:         codec,
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

	var expectedFees sdk.Coins // All the fees which need to be paid for the given tx. includes min consensus fee + every contract flat fee
	gasUnitPrice, found := mfd.rewardsKeeper.GetMinConsensusFee(ctx)
	if found {
		// Estimate the minimum fee expected
		// We use RoundInt here since minimum fee must be GTE calculated amount
		txGasLimit := pkg.NewDecFromUint64(feeTx.GetGas())
		if txGasLimit.IsZero() {
			return ctx, sdkErrors.Wrap(sdkErrors.ErrInvalidRequest, "tx gas limit is not set")
		}
		minFeeExpected := sdk.Coin{
			Denom:  gasUnitPrice.Denom,
			Amount: gasUnitPrice.Amount.Mul(txGasLimit).RoundInt(),
		}
		expectedFees = expectedFees.Add(minFeeExpected)
	}

	// Get flatfees for any contracts being called in the tx.msgs
	for _, m := range tx.GetMsgs() {
		flatFees, _, err := GetContractFlatFees(ctx, mfd.rewardsKeeper, mfd.codec, m)
		if err != nil {
			return ctx, err
		}
		expectedFees = expectedFees.Add(flatFees...)
	}

	txFees := feeTx.GetFee()
	if expectedFees.IsZero() || txFees.IsAllGTE(expectedFees) {
		return next(ctx, tx, simulate)
	}
	return ctx, sdkErrors.Wrapf(sdkErrors.ErrInsufficientFee, "tx fee %s is less than min fee: %s", txFees, expectedFees.String())
}
