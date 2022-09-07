package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.AnteDecorator = TxGasTrackingDecorator{}

// TrackingKeeperExpected defines the expected interface of the TrackingKeeper.
type TrackingKeeperExpected interface {
	TrackNewTx(ctx sdk.Context)
}

// TxGasTrackingDecorator is an Ante decorator that starts the gas tracking for a new transaction.
type TxGasTrackingDecorator struct {
	keeper TrackingKeeperExpected
}

// AnteHandle implements the AnteDecorator interface.
func (d TxGasTrackingDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	d.keeper.TrackNewTx(ctx)

	return next(ctx, tx, simulate)
}

// NewTxGasTrackingDecorator returns a new TxGasTrackingDecorator instance.
func NewTxGasTrackingDecorator(keeper TrackingKeeperExpected) TxGasTrackingDecorator {
	return TxGasTrackingDecorator{keeper: keeper}
}
