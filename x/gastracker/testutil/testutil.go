package testutil

import (
	"crypto/rand"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	gstTypes "github.com/archway-network/archway/x/gastracker"
	"github.com/archway-network/archway/x/gastracker/keeper"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/stretchr/testify/require"
	tmLog "github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	db "github.com/tendermint/tm-db"
	"os"
	"testing"
	"time"
)

var storeKey = sdk.NewKVStoreKey(gstTypes.StoreKey)

func createTestBaseKeeperAndContext(t *testing.T, contractAdmin sdk.AccAddress) (sdk.Context, keeper.Keeper) {
	encodingConfig := simapp.MakeTestEncodingConfig()
	appCodec := encodingConfig.Marshaler

	memDB := db.NewMemDB()
	ms := store.NewCommitMultiStore(memDB)

	mkey := sdk.NewKVStoreKey("test")
	tkey := sdk.NewTransientStoreKey("transient_test")
	tstoreKey := sdk.NewTransientStoreKey(gstTypes.TStoreKey)

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

func GenerateRandomAccAddress() sdk.AccAddress {
	var address sdk.AccAddress = make([]byte, 20)
	_, err := rand.Read(address)
	if err != nil {
		panic(err)
	}
	return address
}

func CreateTestBlockEntry(ctx sdk.Context, k keeper.Keeper, blockTracking gstTypes.BlockGasTracking) {
	k.TrackNewBlock(ctx)
	for _, tx := range blockTracking.TxTrackingInfos {
		k.TrackNewTx(ctx, tx.MaxContractRewards, tx.MaxGasAllowed)
		for _, op := range tx.ContractTrackingInfos {
			addr, _ := sdk.AccAddressFromBech32(op.Address)
			k.TrackContractGasUsage(ctx, addr, wasmTypes.GasConsumptionInfo{
				VMGas:  op.OriginalVmGas,
				SDKGas: op.OriginalSdkGas,
			}, op.Operation)
		}
	}
}

type mockMinter struct{}

func (t mockMinter) GetParams(_ sdk.Context) (params mintTypes.Params) {
	return mintTypes.Params{
		MintDenom:     "test",
		BlocksPerYear: 100,
	}
}

func (t mockMinter) GetMinter(_ sdk.Context) (minter mintTypes.Minter) {
	return mintTypes.Minter{
		AnnualProvisions: sdk.NewDec(76500),
	}
}
