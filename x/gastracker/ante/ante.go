package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/archway-network/archway/x/gastracker/keeper"
)

type TxGasTrackingDecorator struct {
	gasTrackingKeeper keeper.GasTrackingKeeper
}

func (t TxGasTrackingDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if ctx.BlockHeight() <= 1 {
		return next(ctx, tx, simulate)
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()

	feeCoins = feeCoins.Sort()
	rewardCoins := make([]*sdk.DecCoin, len(feeCoins))
	remainingFeeCoins := make([]*sdk.DecCoin, len(feeCoins))

	for i, coin := range feeCoins {
		decCoin := sdk.NewDecCoinFromCoin(coin)
		reward := decCoin.Sub(sdk.NewDecCoinFromDec(coin.Denom, decCoin.Amount.Quo(sdk.NewDec(2))))
		rewardCoins[i] = &reward

		remainingFeeCoin := decCoin.Sub(reward)
		remainingFeeCoins[i] = &remainingFeeCoin
	}

	err = t.gasTrackingKeeper.TrackNewTx(ctx, rewardCoins, remainingFeeCoins, feeTx.GetGas())
	if err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

func NewTxGasTrackingDecorator(gasTrackingKeeper keeper.GasTrackingKeeper) TxGasTrackingDecorator {
	return TxGasTrackingDecorator{gasTrackingKeeper: gasTrackingKeeper}
}
