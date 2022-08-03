package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/archway-network/archway/pkg"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

var _ sdk.AnteDecorator = DeductFeeDecorator{}

// TxFeeRewardsKeeperExpected defines the expected interface for the x/rewards keeper.
type TxFeeRewardsKeeperExpected interface {
	TxFeeRebateRatio(ctx sdk.Context) sdk.Dec
	TrackFeeRebatesRewards(ctx sdk.Context, rewards sdk.Coins)
}

// DeductFeeDecorator deducts fees from the first signer of the tx.
// If the first signer does not have the funds to pay for the fees, return with InsufficientFunds error.
// Call next AnteHandler if fees successfully deducted.
// CONTRACT: Tx must implement FeeTx interface to use DeductFeeDecorator.
type DeductFeeDecorator struct {
	ak             ante.AccountKeeper
	bankKeeper     authTypes.BankKeeper
	feegrantKeeper ante.FeegrantKeeper
	rewardsKeeper  TxFeeRewardsKeeperExpected
}

// NewDeductFeeDecorator returns a new DeductFeeDecorator instance.
func NewDeductFeeDecorator(ak ante.AccountKeeper, bk authTypes.BankKeeper, fk ante.FeegrantKeeper, rk TxFeeRewardsKeeperExpected) DeductFeeDecorator {
	return DeductFeeDecorator{
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
		return ctx, sdkErrors.Wrap(sdkErrors.ErrTxDecode, "Tx must be a FeeTx")
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
			return ctx, sdkErrors.Wrap(sdkErrors.ErrInvalidRequest, "fee grants are not enabled")
		}

		if !feeGranter.Equals(feePayer) {
			if err := dfd.feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, fee, tx.GetMsgs()); err != nil {
				return ctx, sdkErrors.Wrapf(err, "%s not allowed to pay fees from %s", feeGranter, feePayer)
			}
		}

		deductFeesFrom = feeGranter
	}

	deductFeesFromAcc := dfd.ak.GetAccount(ctx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return ctx, sdkErrors.Wrapf(sdkErrors.ErrUnknownAddress, "fee payer address (%s) does not exist", deductFeesFrom)
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
		return sdkErrors.Wrapf(sdkErrors.ErrInsufficientFee, "invalid fee amount: %s", fees)
	}

	// Check if transaction has wasmd operations
	hasWasmMsgs := false
	for _, msg := range tx.GetMsgs() {
		// We can use switch here, but breaking the for loop from switch is less readable
		if _, ok := msg.(*wasmdTypes.MsgExecuteContract); ok {
			hasWasmMsgs = true
			break
		}
		if _, ok := msg.(*wasmdTypes.MsgMigrateContract); ok {
			hasWasmMsgs = true
			break
		}
	}

	// Send everything to the fee collector account if rewards are disabled or transaction is not wasm related
	rebateRatio := dfd.rewardsKeeper.TxFeeRebateRatio(ctx)
	if rebateRatio.IsZero() || !hasWasmMsgs {
		if err := dfd.bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), authTypes.FeeCollectorName, fees); err != nil {
			return sdkErrors.Wrapf(sdkErrors.ErrInsufficientFunds, err.Error())
		}
		return nil
	}

	// Split the fees between the fee collector account and the rewards collector account
	rewardsFees, authFees := pkg.SplitCoins(fees, rebateRatio)

	if !authFees.Empty() {
		if err := dfd.bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), authTypes.FeeCollectorName, authFees); err != nil {
			return sdkErrors.Wrapf(sdkErrors.ErrInsufficientFunds, err.Error())
		}
	}

	if !rewardsFees.Empty() {
		if err := dfd.bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), rewardsTypes.ContractRewardCollector, rewardsFees); err != nil {
			return sdkErrors.Wrapf(sdkErrors.ErrInsufficientFunds, err.Error())
		}
	}

	// Track transaction fee rewards
	dfd.rewardsKeeper.TrackFeeRebatesRewards(ctx, rewardsFees)

	return nil
}
