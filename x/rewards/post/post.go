package post

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.PostDecorator = ReimburseFlatFees{}

func NewReimburseFlatFees(rk RewardsKeeper) ReimburseFlatFees {
	return ReimburseFlatFees{rk: rk}
}

type RewardsKeeper interface {
	MaybeReimburseFlatFees(ctx sdk.Context, txSuccess bool, feePayer sdk.AccAddress) (reimbursed sdk.Coins, err error)
}

type ReimburseFlatFees struct {
	rk RewardsKeeper
}

func (r ReimburseFlatFees) PostHandle(ctx sdk.Context, tx sdk.Tx, simulate, success bool, next sdk.PostHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkErrors.ErrTxDecode, "Tx must be a FeeTx")
	}
	feeGranter := feeTx.FeeGranter()
	feePayer := feeTx.FeePayer()
	if feeGranter != nil {
		feePayer = feeGranter
	}
	_, err = r.rk.MaybeReimburseFlatFees(ctx, success, feePayer)
	if err != nil {
		return ctx, err
	}
	return next(ctx, tx, simulate, success)
}
