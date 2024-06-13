package testutils

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	wasmdtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/archway-network/archway/x/tracking/keeper"
	"github.com/archway-network/archway/x/tracking/types"
)

func TrackingKeeper(tb testing.TB) (keeper.Keeper, sdk.Context) {
	tb.Helper()
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey("m_tracking")
	tStoreKey := storetypes.NewTransientStoreKey("t_tracking")

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewTestLogger(tb), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	stateStore.MountStoreWithDB(tStoreKey, storetypes.StoreTypeTransient, db)
	require.NoError(tb, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		wasmdtypes.NewDefaultWasmGasRegister(),
		log.NewTestLogger(tb),
	)
	ctx := sdk.NewContext(stateStore, tmproto.Header{
		Height: 1,
	}, false, log.NewNopLogger())

	return k, ctx
}

type MockTrackingKeeper struct {
	GetCurrentTxIDFn          func(ctx sdk.Context) uint64
	GetBlockTrackingInfoFn    func(ctx sdk.Context, height int64) types.BlockTracking
	RemoveBlockTrackingInfoFn func(ctx sdk.Context, height int64)
}

func (m MockTrackingKeeper) GetCurrentTxID(ctx sdk.Context) uint64 {
	return m.GetCurrentTxIDFn(ctx)
}

func (m MockTrackingKeeper) GetBlockTrackingInfo(ctx sdk.Context, height int64) types.BlockTracking {
	return m.GetBlockTrackingInfoFn(ctx, height)
}

func (m MockTrackingKeeper) RemoveBlockTrackingInfo(ctx sdk.Context, height int64) {
	m.RemoveBlockTrackingInfoFn(ctx, height)
}
