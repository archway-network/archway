package gastracker

import (
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type dummyTx struct {
	Gas uint64
	Fee sdk.Coins
}

func (d dummyTx) GetMsgs() []sdk.Msg {
	panic("should not be invoked by AnteHandler")
}

func (d dummyTx) ValidateBasic() error {
	panic("should not be invoked by AnteHandler")
}

func (d dummyTx) GetGas() uint64 {
	return d.Gas
}

func (d dummyTx) GetFee() sdk.Coins {
	return d.Fee
}

func (d dummyTx) FeePayer() sdk.AccAddress {
	panic("should not be invoked by AnteHandler")
}

func (d dummyTx) FeeGranter() sdk.AccAddress {
	panic("should not be invoked by AnteHandler")
}

type InvalidTx struct {
}

func (i InvalidTx) GetMsgs() []sdk.Msg {
	panic("not implemented")
}

func (i InvalidTx) ValidateBasic() error {
	panic("not implemented")
}

func dummyNextAnteHandler(_ sdk.Context, _ sdk.Tx, _ bool) (newCtx sdk.Context, err error) {
	return sdk.Context{}, nil
}

func TestGasTrackingAnteHandler(t *testing.T) {
	ctx, keeper := CreateTestKeeperAndContext(t)

	testTxGasTrackingDecorator := NewTxGasTrackingDecorator(keeper)

	_, err := testTxGasTrackingDecorator.AnteHandle(ctx, &InvalidTx{}, false, dummyNextAnteHandler)
	assert.EqualError(
		t,
		err,
		sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx").Error(),
		"Gastracking ante handler should return expected error",
	)

	_, err = testTxGasTrackingDecorator.AnteHandle(ctx.WithBlockHeight(1), &InvalidTx{}, false, dummyNextAnteHandler)
	assert.NoError(t, err, "Ante handler should not do anything for blockheight less then or equal to 1")

	testTx := dummyTx{
		Gas: 500,
		Fee: sdk.NewCoins(sdk.NewInt64Coin("test", 10)),
	}

	expectedDecCoins := make([]*sdk.DecCoin, len(testTx.Fee))
	for i, coin := range testTx.Fee {
		expectedDecCoins[i] = new(sdk.DecCoin)
		*expectedDecCoins[i] = sdk.NewDecCoinFromCoin(sdk.NewCoin(coin.Denom, coin.Amount.QuoRaw(2)))
	}

	_, err = testTxGasTrackingDecorator.AnteHandle(ctx, testTx, false, dummyNextAnteHandler)
	assert.EqualError(
		t,
		err,
		gstTypes.ErrBlockTrackingDataNotFound.Error(),
		"Gastracking ante handler should return expected error",
	)

	err = keeper.TrackNewBlock(ctx)
	assert.NoError(t, err, "New block gas tracking should succeed")

	_, err = testTxGasTrackingDecorator.AnteHandle(ctx, testTx, false, dummyNextAnteHandler)
	assert.NoError(
		t,
		err,
		"Gastracking ante handler should not return an error",
	)

	currentBlockTrackingInfo, err := keeper.GetCurrentBlockTrackingInfo(ctx)
	assert.NoError(t, err, "Current block tracking info should exists")

	assert.Equal(t, 1, len(currentBlockTrackingInfo.TxTrackingInfos), "Only 1 txtracking info should be there")
	assert.Equal(t, testTx.Gas, currentBlockTrackingInfo.TxTrackingInfos[0].MaxGasAllowed, "MaxGasAllowed must match the Gas field of tx")
	assert.Equal(t, expectedDecCoins, currentBlockTrackingInfo.TxTrackingInfos[0].MaxContractRewards, "MaxContractReward must be half of the tx fees")

	testTx = dummyTx{
		Gas: 100,
		Fee: sdk.NewCoins(sdk.NewInt64Coin("test", 20)),
	}

	expectedDecCoins = make([]*sdk.DecCoin, len(testTx.Fee))
	for i, coin := range testTx.Fee {
		expectedDecCoins[i] = new(sdk.DecCoin)
		*expectedDecCoins[i] = sdk.NewDecCoinFromCoin(sdk.NewCoin(coin.Denom, coin.Amount.QuoRaw(2)))
	}

	_, err = testTxGasTrackingDecorator.AnteHandle(ctx, testTx, false, dummyNextAnteHandler)
	assert.NoError(
		t,
		err,
		"Gastracking ante handler should not return an error",
	)

	currentBlockTrackingInfo, err = keeper.GetCurrentBlockTrackingInfo(ctx)
	assert.NoError(t, err, "Current block tracking info should exists")

	assert.Equal(t, 2, len(currentBlockTrackingInfo.TxTrackingInfos), "Only 1 txtracking info should be there")
	assert.Equal(t, testTx.Gas, currentBlockTrackingInfo.TxTrackingInfos[1].MaxGasAllowed, "MaxGasAllowed must match the Gas field of tx")
	assert.Equal(t, expectedDecCoins, currentBlockTrackingInfo.TxTrackingInfos[1].MaxContractRewards, "MaxContractReward must be half of the tx fees")
}
