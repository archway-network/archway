package ante

import (
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
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
	rewardsKeeper RewardsFeeReaderExpected
}

// NewMinFeeDecorator returns a new MinFeeDecorator instance.
func NewMinFeeDecorator(rk RewardsFeeReaderExpected) MinFeeDecorator {
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

	for _, m := range tx.GetMsgs() {
		contractAddr, cwExecMsgFound := getMsgContractAddress(m)
		if cwExecMsgFound {
			ca := sdk.MustAccAddressFromBech32(contractAddr)
			flatfee, found := mfd.rewardsKeeper.GetFlatFee(ctx, ca)
			if found {
				panic(flatfee)
			}
		}
	}

	// Check (skip if the expected amount is zero)
	if minFeeExpected.Amount.IsZero() || txFees.IsAnyGTE(sdk.Coins{minFeeExpected}) {
		return next(ctx, tx, simulate)
	}

	return ctx, sdkErrors.Wrapf(sdkErrors.ErrInsufficientFee, "tx fee %s is less than min fee: %s", txFees, minFeeExpected)
}

func getMsgContractAddress(m sdk.Msg) (contractAddress string, isContractExecute bool) {
	switch msg := m.(type) {
	case *wasmTypes.MsgExecuteContract:
		{
			return msg.Contract, true
		}
	case *authz.MsgExec:
		{
			// todo: decode authz msgs and check if they are execute contract
			// var wrappedMsg sdk.Msg
			// err := dec.codec.UnpackAny(msg, &wrappedMsg)
			// if err != nil {
			// 	return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "error decoding authz messages")
			// }
			return "", false
		}
	default:
		return "", false
	}
}
