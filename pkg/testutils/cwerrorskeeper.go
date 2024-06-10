package testutils

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	"github.com/archway-network/archway/x/cwerrors/keeper"
	"github.com/archway-network/archway/x/cwerrors/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func CWErrorsKeeper(tb testing.TB) (keeper.Keeper, sdk.Context) {
	tb.Helper()
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey("m_cwerrors")
	tStoreKey := storetypes.NewTransientStoreKey(types.TStoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewTestLogger(tb), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	stateStore.MountStoreWithDB(tStoreKey, storetypes.StoreTypeTransient, db)
	require.NoError(tb, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	bankKeeper := MockBankKeeper{
		SendCoinsFromAccountToModuleFn: func(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
			return nil
		},
	}
	k := keeper.NewKeeper(
		cdc,
		storeKey,
		tStoreKey,
		nil,
		bankKeeper,
		nil,
		"cosmos1a48wdtjn3egw7swhfkeshwdtjvs6hq9nlyrwut", // random addr for gov module
	)
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	params := types.DefaultParams()
	_ = k.SetParams(ctx, params)

	return k, ctx
}

type MockBankKeeper struct {
	SendCoinsFromAccountToModuleFn func(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccountFn func(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModuleFn  func(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	BlockedAddrFn                  func(addr sdk.AccAddress) bool
}

func (k MockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	if k.SendCoinsFromAccountToModuleFn == nil {
		panic("not supposed to be called!")
	}
	return k.SendCoinsFromAccountToModuleFn(ctx, senderAddr, recipientModule, amt)
}

func (k MockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if k.SendCoinsFromAccountToModuleFn == nil {
		panic("not supposed to be called!")
	}
	return k.SendCoinsFromModuleToAccountFn(ctx, senderModule, recipientAddr, amt)
}

func (k MockBankKeeper) SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error {
	if k.SendCoinsFromAccountToModuleFn == nil {
		panic("not supposed to be called!")
	}
	return k.SendCoinsFromModuleToModuleFn(ctx, senderModule, recipientModule, amt)
}

func (k MockBankKeeper) BlockedAddr(addr sdk.AccAddress) bool {
	if k.SendCoinsFromAccountToModuleFn == nil {
		panic("not supposed to be called!")
	}
	return k.BlockedAddrFn(addr)
}
