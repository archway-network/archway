package ante

import (
	"os"
	"testing"
	"time"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmLog "github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	db "github.com/tendermint/tm-db"

	"github.com/archway-network/archway/x/gastracker"
	gstTypes "github.com/archway-network/archway/x/gastracker"
	"github.com/archway-network/archway/x/gastracker/keeper"
)

// NOTE: this is needed to allow the keeper to set BlockGasTracking
var (
	storeKey = sdk.NewKVStoreKey(gastracker.StoreKey)
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
	ctx, keeper := CreateTestKeeperAndContext(t, sdk.AccAddress{})

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

	expectedDecCoins := make([]sdk.DecCoin, len(testTx.Fee))
	for i, coin := range testTx.Fee {
		expectedDecCoins[i] = sdk.NewDecCoinFromCoin(sdk.NewCoin(coin.Denom, coin.Amount.QuoRaw(2)))
	}

	keeper.TrackNewBlock(ctx)

	_, err = testTxGasTrackingDecorator.AnteHandle(ctx, testTx, false, dummyNextAnteHandler)
	assert.NoError(
		t,
		err,
		"Gastracking ante handler should not return an error",
	)

	currentBlockTrackingInfo := keeper.GetCurrentBlockTracking(ctx)
	assert.NoError(t, err, "Current block tracking info should exists")

	assert.Equal(t, 1, len(currentBlockTrackingInfo.TxTrackingInfos), "Only 1 txtracking info should be there")
	assert.Equal(t, testTx.Gas, currentBlockTrackingInfo.TxTrackingInfos[0].MaxGasAllowed, "MaxGasAllowed must match the Gas field of tx")
	assert.Equal(t, expectedDecCoins, currentBlockTrackingInfo.TxTrackingInfos[0].MaxContractRewards, "MaxContractReward must be half of the tx fees")

	testTx = dummyTx{
		Gas: 100,
		Fee: sdk.NewCoins(sdk.NewInt64Coin("test", 20)),
	}

	expectedDecCoins = make([]sdk.DecCoin, len(testTx.Fee))
	for i, coin := range testTx.Fee {
		expectedDecCoins[i] = sdk.NewDecCoinFromCoin(sdk.NewCoin(coin.Denom, coin.Amount.QuoRaw(2)))
	}

	_, err = testTxGasTrackingDecorator.AnteHandle(ctx, testTx, false, dummyNextAnteHandler)
	assert.NoError(
		t,
		err,
		"Gastracking ante handler should not return an error",
	)

	currentBlockTrackingInfo = keeper.GetCurrentBlockTracking(ctx)

	assert.Equal(t, 2, len(currentBlockTrackingInfo.TxTrackingInfos), "Only 1 txtracking info should be there")
	assert.Equal(t, testTx.Gas, currentBlockTrackingInfo.TxTrackingInfos[1].MaxGasAllowed, "MaxGasAllowed must match the Gas field of tx")
	assert.Equal(t, expectedDecCoins, currentBlockTrackingInfo.TxTrackingInfos[1].MaxContractRewards, "MaxContractReward must be half of the tx fees")
}

// TODO: this is shared test util, that is copied
// from /keeper/keeper_test, refactor
func createTestBaseKeeperAndContext(t *testing.T, contractAdmin sdk.AccAddress) (sdk.Context, keeper.Keeper) {
	encodingConfig := simapp.MakeTestEncodingConfig()
	appCodec := encodingConfig.Marshaler

	memDB := db.NewMemDB()
	ms := store.NewCommitMultiStore(memDB)

	mkey := sdk.NewKVStoreKey("test")
	tkey := sdk.NewTransientStoreKey("transient_test")
	tstoreKey := sdk.NewTransientStoreKey(gastracker.TStoreKey)

	ms.MountStoreWithDB(mkey, sdk.StoreTypeIAVL, memDB)
	ms.MountStoreWithDB(tkey, sdk.StoreTypeIAVL, memDB)
	ms.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, memDB)
	ms.MountStoreWithDB(tstoreKey, sdk.StoreTypeTransient, memDB)

	err := ms.LoadLatestVersion()
	require.NoError(t, err, "Loading latest version should not fail")

	pkeeper := paramskeeper.NewKeeper(appCodec, encodingConfig.Amino, mkey, tkey)
	subspace := pkeeper.Subspace(gstTypes.ModuleName)

	keeper := keeper.NewGasTrackingKeeper(
		storeKey,
		appCodec,
		subspace,
		NewTestContractInfoView(contractAdmin.String()),
		wasmkeeper.NewDefaultWasmGasRegister(),
		nil,
	)

	ctx := sdk.NewContext(ms, tmproto.Header{
		Height: 10000,
		Time:   time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
	}, false, tmLog.NewTMLogger(os.Stdout))

	params := gstTypes.DefaultParams()
	subspace.SetParamSet(ctx, &params)
	return ctx, keeper
}

func CreateTestKeeperAndContext(t *testing.T, contractAdmin sdk.AccAddress) (sdk.Context, keeper.Keeper) {
	return createTestBaseKeeperAndContext(t, contractAdmin)
}

type TestContractInfoView struct {
	keeper.ContractInfoView
	adminMap     map[string]string
	defaultAdmin string
}

func NewTestContractInfoView(defaultAdmin string) *TestContractInfoView {
	return &TestContractInfoView{
		adminMap:     make(map[string]string),
		defaultAdmin: defaultAdmin,
	}
}

func (t *TestContractInfoView) GetContractInfo(_ sdk.Context, contractAddress sdk.AccAddress) *wasmTypes.ContractInfo {
	if admin, ok := t.adminMap[contractAddress.String()]; ok {
		return &wasmTypes.ContractInfo{Admin: admin}
	} else {
		return &wasmTypes.ContractInfo{Admin: t.defaultAdmin}
	}
}

func (t *TestContractInfoView) AddContractToAdminMapping(contractAddress string, admin string) {
	t.adminMap[contractAddress] = admin
}

var _ keeper.ContractInfoView = &TestContractInfoView{}
