package ante

import (
	"bytes"
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/archway-network/archway/pkg"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

var _ sdk.AnteDecorator = DeductFeeDecorator{}

type BankKeeper interface {
	authTypes.BankKeeper
	BurnCoins(context.Context, string, sdk.Coins) error
}

type CWFeesKeeper interface {
	IsGrantingContract(ctx context.Context, granter sdk.AccAddress) (bool, error)
	RequestGrant(ctx context.Context, grantingContract sdk.AccAddress, txMsgs []sdk.Msg, wantFees sdk.Coins, signers []sdk.AccAddress) error
}

// DeductFeeDecorator deducts fees from the first signer of the tx.
// If the first signer does not have the funds to pay for the fees, return with InsufficientFunds error.
// Call next AnteHandler if fees successfully deducted.
// CONTRACT: Tx must implement FeeTx interface to use DeductFeeDecorator.
type DeductFeeDecorator struct {
	codec          codec.BinaryCodec
	ak             ante.AccountKeeper
	bankKeeper     BankKeeper
	feegrantKeeper ante.FeegrantKeeper
	rewardsKeeper  RewardsKeeperExpected
	cwFeesKeeper   CWFeesKeeper
}

// NewDeductFeeDecorator returns a new DeductFeeDecorator instance.
func NewDeductFeeDecorator(codec codec.BinaryCodec, ak ante.AccountKeeper, bk BankKeeper, fk ante.FeegrantKeeper, rk RewardsKeeperExpected, ck CWFeesKeeper) DeductFeeDecorator {
	return DeductFeeDecorator{
		codec:          codec,
		ak:             ak,
		bankKeeper:     bk,
		feegrantKeeper: fk,
		rewardsKeeper:  rk,
		cwFeesKeeper:   ck,
	}
}

// AnteHandle implements the ante.AnteDecorator interface.
func (dfd DeductFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkErrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if addr := dfd.ak.GetModuleAddress(authTypes.FeeCollectorName); addr == nil {
		return ctx, fmt.Errorf("fee collector module account (%s) has not been set", authTypes.FeeCollectorName)
	}

	deductFeesFrom, err := dfd.getFeePayer(ctx, feeTx)
	if err != nil {
		return ctx, err
	}

	deductFeesFromAcc := dfd.ak.GetAccount(ctx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return ctx, errorsmod.Wrapf(sdkErrors.ErrUnknownAddress, "fee payer address (%s) does not exist", deductFeesFrom)
	}

	// Deduct the fees
	if !feeTx.GetFee().IsZero() {
		if err := dfd.deductFees(ctx, tx, deductFeesFromAcc, feeTx.GetFee()); err != nil {
			return ctx, err
		}
	}

	events := sdk.Events{sdk.NewEvent(sdk.EventTypeTx,
		sdk.NewAttribute(sdk.AttributeKeyFee, feeTx.GetFee().String()),
	)}
	ctx.EventManager().EmitEvents(events)

	return next(ctx, tx, simulate)
}

// getFeePayer returns the address of the entity will we get fees from.
func (dfd DeductFeeDecorator) getFeePayer(ctx sdk.Context, tx sdk.Tx) (payer sdk.AccAddress, err error) {
	feeTx, _ := tx.(sdk.FeeTx)
	payer = feeTx.FeePayer()
	granter := feeTx.FeeGranter()
	// in case granter is nil or payer and granter are equal
	// then we just return the fee payer as the entity who pays the fees.
	if granter == nil || bytes.Equal(payer.Bytes(), granter) {
		return payer, nil
	}

	switch {
	// we check x/cwfees first
	case dfd.cwFeesKeeper != nil:
		isCWGranter, err := dfd.cwFeesKeeper.IsGrantingContract(ctx, granter)
		if err != nil {
			return nil, err
		}
		// the contract is a cw granter, so we request fees from it.
		if isCWGranter {
			sigTx, ok := tx.(authsigning.SigVerifiableTx)
			if !ok {
				return nil, errorsmod.Wrap(sdkErrors.ErrTxDecode, "Tx must be a SigVerifiableTx")
			}
			signers, err := sigTx.GetSigners()
			if err != nil {
				return nil, errorsmod.Wrap(sdkErrors.ErrInvalidRequest, "cannot get signers from tx")
			}
			var signerAddrs []sdk.AccAddress
			for _, s := range signers {
				signerAddrs = append(signerAddrs, sdk.AccAddress(s))
			}
			err = dfd.cwFeesKeeper.RequestGrant(ctx, granter, feeTx.GetMsgs(), feeTx.GetFee(), signerAddrs)
			if err != nil {
				return nil, errorsmod.Wrapf(err, "%s contract is not allowed to pay fees from %s", granter, payer)
			}
			return granter, nil
		}
		// cannot be handled through x/cwfees, let's try with x/feegrant
		fallthrough

	// we check x/feegrant
	case dfd.feegrantKeeper != nil:
		err = dfd.feegrantKeeper.UseGrantedFees(ctx, granter, payer, feeTx.GetFee(), feeTx.GetMsgs())
		if err != nil {
			return nil, errorsmod.Wrapf(err, "%s not allowed to pay fees from %s", granter, payer)
		}
		return granter, nil
	// the default case is
	default:
		return nil, errorsmod.Wrap(sdkErrors.ErrInvalidRequest, "fee grants are not enabled")
	}
}

// deductFees deducts fees from the given account if rewards calculation and distribution is enabled.
// If rewards module is disabled, all the fees are sent to the fee collector account.
// NOTE: this is the only logic being changed.
func (dfd DeductFeeDecorator) deductFees(ctx sdk.Context, tx sdk.Tx, acc sdk.AccountI, fees sdk.Coins) error {
	if !fees.IsValid() {
		return errorsmod.Wrapf(sdkErrors.ErrInsufficientFee, "invalid fee amount: %s", fees)
	}

	var flatFees sdk.Coins
	// Check if transaction has wasmd operations
	hasWasmMsgs := false
	for _, m := range tx.GetMsgs() {
		contractFlatFees, hwm, err := GetContractFlatFees(ctx, dfd.rewardsKeeper, dfd.codec, m)
		if err != nil {
			return err
		}
		// set hasWasmMsgs, if it is still false;
		if !hasWasmMsgs {
			hasWasmMsgs = hwm
		}
		for _, cff := range contractFlatFees {
			flatFees = flatFees.Add(cff.FlatFees...)
		}
	}

	// Send everything to the fee collector account if rewards are disabled or transaction is not wasm related
	rebateRatio := dfd.rewardsKeeper.TxFeeRebateRatio(ctx)
	if rebateRatio.IsZero() || !hasWasmMsgs {
		if err := dfd.bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), authTypes.FeeCollectorName, fees); err != nil {
			return errorsmod.Wrapf(sdkErrors.ErrInsufficientFunds, err.Error())
		}
		return nil
	}

	if !flatFees.Empty() {
		if err := dfd.bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), rewardsTypes.ContractRewardCollector, flatFees); err != nil {
			return errorsmod.Wrapf(sdkErrors.ErrInsufficientFunds, err.Error())
		}
		fees = fees.Sub(flatFees...) // reduce flatfees from the sent fees amount
	}

	// Split the fees between the fee collector account and the rewards collector account
	rewardsFees, authFees := pkg.SplitCoins(fees, rebateRatio)

	if !authFees.Empty() {
		if err := dfd.bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), authTypes.FeeCollectorName, authFees); err != nil {
			return errorsmod.Wrapf(sdkErrors.ErrInsufficientFunds, err.Error())
		}
		// burn the auth fees.
		if err := dfd.bankKeeper.BurnCoins(ctx, authTypes.FeeCollectorName, authFees); err != nil {
			return errorsmod.Wrapf(sdkErrors.ErrInsufficientFunds, err.Error())
		}
	}

	if !rewardsFees.Empty() {
		if err := dfd.bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), rewardsTypes.ContractRewardCollector, rewardsFees); err != nil {
			return errorsmod.Wrapf(sdkErrors.ErrInsufficientFunds, err.Error())
		}
	}

	// Track transaction fee rewards
	dfd.rewardsKeeper.TrackFeeRebatesRewards(ctx, rewardsFees)

	return nil
}
