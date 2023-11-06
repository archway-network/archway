package ante

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/archway-network/archway/pkg"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

var _ sdk.AnteDecorator = DeductFeeDecorator{}

type BankKeeper interface {
	authTypes.BankKeeper
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
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
}

// NewDeductFeeDecorator returns a new DeductFeeDecorator instance.
func NewDeductFeeDecorator(codec codec.BinaryCodec, ak ante.AccountKeeper, bk BankKeeper, fk ante.FeegrantKeeper, rk RewardsKeeperExpected) DeductFeeDecorator {
	return DeductFeeDecorator{
		codec:          codec,
		ak:             ak,
		bankKeeper:     bk,
		feegrantKeeper: fk,
		rewardsKeeper:  rk,
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

	fee := feeTx.GetFee()
	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()

	deductFeesFrom := feePayer

	// If feegranter set, deduct fee from feegranter account (only when feegrant is enabled)
	if feeGranter != nil {
		if dfd.feegrantKeeper == nil {
			return ctx, errorsmod.Wrap(sdkErrors.ErrInvalidRequest, "fee grants are not enabled")
		}

		if !feeGranter.Equals(feePayer) {
			if err := dfd.feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, fee, tx.GetMsgs()); err != nil {
				return ctx, errorsmod.Wrapf(err, "%s not allowed to pay fees from %s", feeGranter, feePayer)
			}
		}

		deductFeesFrom = feeGranter
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

// deductFees deducts fees from the given account if rewards calculation and distribution is enabled.
// If rewards module is disabled, all the fees are sent to the fee collector account.
// NOTE: this is the only logic being changed.
func (dfd DeductFeeDecorator) deductFees(ctx sdk.Context, tx sdk.Tx, acc authTypes.AccountI, fees sdk.Coins) error {
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
