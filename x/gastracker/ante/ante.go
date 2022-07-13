package ante

import (
	"github.com/archway-network/archway/x/gastracker"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type GasTrackingKeeper interface {
	GetParams(ctx sdk.Context) gastracker.Params
	TrackNewTx(ctx sdk.Context, rewardCoins []*sdk.DecCoin, gas uint64) error
}

type TxGasTrackingDecorator struct {
	gasTrackingKeeper GasTrackingKeeper
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

	params := t.gasTrackingKeeper.GetParams(ctx)
	for i, coin := range feeCoins {
		decCoin := sdk.NewDecCoinFromCoin(coin)
		reward := decCoin.Sub(sdk.NewDecCoinFromDec(coin.Denom, decCoin.Amount.Mul(params.DappTxFeeRebateRatio)))
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
