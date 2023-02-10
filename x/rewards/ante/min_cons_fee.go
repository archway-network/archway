package ante

import (
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/archway-network/archway/pkg"
)

// RewardsFeeReaderExpected defines the expected interface for the x/rewards keeper.
type RewardsFeeReaderExpected interface {
	GetMinConsensusFee(ctx sdk.Context) (sdk.DecCoin, bool)
	GetFlatFee(ctx sdk.Context, contractAddr sdk.AccAddress) (sdk.Coin, bool)
}

// MinFeeDecorator rejects transaction if its fees are less than minimum fees defined by the x/rewards module.
// Estimation is done using the minimum consensus fee value which is the minimum gas unit price.
// The minimum consensus fee value is defined by block dApp rewards and rewards distribution parameters.
// CONTRACT: Tx must implement FeeTx interface to use MinFeeDecorator.
type MinFeeDecorator struct {
	codec         codec.BinaryCodec
	rewardsKeeper RewardsFeeReaderExpected
}

// NewMinFeeDecorator returns a new MinFeeDecorator instance.
func NewMinFeeDecorator(codec codec.BinaryCodec, rk RewardsFeeReaderExpected) MinFeeDecorator {
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
		flatFees, err := mfd.getContractFlatFees(ctx, m)
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

func (mfd MinFeeDecorator) getContractFlatFees(ctx sdk.Context, m sdk.Msg) (sdk.Coins, error) {
	switch msg := m.(type) {
	case *wasmTypes.MsgExecuteContract: // if msg is contract execute, fetch flatfee for msg.Contract address
		{
			ca, err := sdk.AccAddressFromBech32(msg.Contract)
			if err != nil {
				return nil, err
			}
			fee, found := mfd.rewardsKeeper.GetFlatFee(ctx, ca)
			if found {
				return sdk.NewCoins(fee), nil
			}
		}
	case *authz.MsgExec: // if msg is authz msg, unwrap the msg and check if any are wasmTypes.MsgExecuteContract
		{
			var flatfees sdk.Coins
			for _, v := range msg.Msgs {
				var wrappedMsg sdk.Msg
				err := mfd.codec.UnpackAny(v, &wrappedMsg)
				if err != nil {
					return nil, sdkErrors.Wrapf(sdkErrors.ErrUnauthorized, "error decoding authz messages")
				}
				fees, err := mfd.getContractFlatFees(ctx, wrappedMsg)
				if err != nil {
					return nil, err
				}
				flatfees = flatfees.Add(fees...)
			}
			return flatfees, nil
		}
	}
	return nil, nil
}
