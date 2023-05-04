package ante

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
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
	computationalGasPrice := mfd.calcComputationalFees(ctx, feeTx.GetGas())
	expectedFees = expectedFees.Add(computationalGasPrice)

	// Get flatfees for any contracts being called in the tx.msgs
	for _, m := range tx.GetMsgs() {
		contractFlatFees, _, err := GetContractFlatFees(ctx, mfd.rewardsKeeper, mfd.codec, m)
		if err != nil {
			return ctx, err
		}
		for _, cff := range contractFlatFees {
			mfd.rewardsKeeper.CreateFlatFeeRewardsRecords(ctx, cff.ContractAddress, cff.FlatFees)
			expectedFees = expectedFees.Add(cff.FlatFees...)
		}
	}

	txFees := feeTx.GetFee()
	if expectedFees.IsZero() || txFees.IsAllGTE(expectedFees) {
		return next(ctx, tx, simulate)
	}
	return ctx, sdkErrors.Wrapf(sdkErrors.ErrInsufficientFee, "tx fee %s is less than min fee: %s", txFees, expectedFees.String())
}

func (mfd MinFeeDecorator) calcComputationalFees(ctx sdk.Context, gasLimit uint64) sdk.Coin {
	minPoG := mfd.rewardsKeeper.MinimumPriceOfGas(ctx)
	antiDoSPoG, found := mfd.rewardsKeeper.GetMinConsensusFee(ctx)
	// no anti dos price of gas is set, then we fallback to the minimum price of gas
	if !found {
		return sdk.NewCoin(minPoG.Denom, sdk.NewIntFromUint64(gasLimit).Mul(minPoG.Amount))
	}
	// otherwise we need to compute the fees in both cases and see which is higher
	if minPoG.Denom != antiDoSPoG.Denom {
		panic("conflict between minPoG denom and antiDoSPog denom")
	}

	gasLimitInt := sdk.NewIntFromUint64(gasLimit)
	minPoGAmt := gasLimitInt.Mul(minPoG.Amount)
	minAntiDoSPoG := sdk.NewDecFromInt(gasLimitInt).Mul(antiDoSPoG.Amount).RoundInt()
	return sdk.NewCoin(minPoG.Denom, sdk.MaxInt(minPoGAmt, minAntiDoSPoG))
}
