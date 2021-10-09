package gastracker

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type TxGasTrackingDecorator struct {
	gasTrackingKeeper GasTrackingKeeper
}

func (t TxGasTrackingDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()

	feeCoins = feeCoins.Sort()
	rewardCoins := make([]*sdk.DecCoin, len(feeCoins))

	for i, coin := range feeCoins {
		decCoin := sdk.NewDecCoinFromCoin(coin)
		reward := decCoin.Sub(sdk.NewDecCoinFromDec(coin.Denom, decCoin.Amount.Quo(sdk.NewDec(2))))
		rewardCoins[i] = &reward
	}

	err = t.gasTrackingKeeper.TrackNewTx(ctx, rewardCoins, feeTx.GetGas())
	if err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

func NewTxGasTrackingDecorator(gasTrackingKeeper GasTrackingKeeper) TxGasTrackingDecorator {
	return TxGasTrackingDecorator{gasTrackingKeeper: gasTrackingKeeper}
}