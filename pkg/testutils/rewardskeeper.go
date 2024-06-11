package testutils

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	"github.com/archway-network/archway/x/rewards/keeper"
	"github.com/archway-network/archway/x/rewards/types"
	rewardstypes "github.com/archway-network/archway/x/rewards/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func RewardsKeeper(tb testing.TB) (keeper.Keeper, sdk.Context, MockBankKeeper, MockContractViewer) {
	tb.Helper()
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey("m_rewards")
	tStoreKey := storetypes.NewTransientStoreKey("t_rewards")

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewTestLogger(tb), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	stateStore.MountStoreWithDB(tStoreKey, storetypes.StoreTypeTransient, db)
	require.NoError(tb, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	wasmKeeper := NewMockContractViewer()
	bankKeeper := MockBankKeeper{}
	k := keeper.NewKeeper(
		cdc,
		storeKey,
		wasmKeeper,
		nil,
		nil,
		bankKeeper,
		"cosmos1a48wdtjn3egw7swhfkeshwdtjvs6hq9nlyrwut", // random addr for gov module
		log.NewTestLogger(tb),
	)
	ctx := sdk.NewContext(stateStore, tmproto.Header{
		Height: 1,
	}, false, log.NewNopLogger())

	return k, ctx, bankKeeper, *wasmKeeper
}

type MockRewardsKeeper struct {
	ComputationalPriceOfGasFn func(ctx sdk.Context) sdk.DecCoin
	GetContractMetadataFn     func(ctx sdk.Context, contractAddr sdk.AccAddress) *rewardstypes.ContractMetadata
}

func (k MockRewardsKeeper) ComputationalPriceOfGas(ctx sdk.Context) sdk.DecCoin {
	if k.ComputationalPriceOfGasFn == nil {
		panic("not supposed to be called!")
	}
	return k.ComputationalPriceOfGasFn(ctx)
}

func (k MockRewardsKeeper) GetContractMetadata(ctx sdk.Context, contractAddr sdk.AccAddress) *rewardstypes.ContractMetadata {
	if k.GetContractMetadataFn == nil {
		panic("not supposed to be called!")
	}
	return k.GetContractMetadataFn(ctx, contractAddr)
}
