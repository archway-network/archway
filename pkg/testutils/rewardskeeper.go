package testutils

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	"github.com/archway-network/archway/x/rewards/keeper"
	"github.com/archway-network/archway/x/rewards/types"
	trackingtypes "github.com/archway-network/archway/x/tracking/types"
)

func RewardsKeeper(tb testing.TB) (keeper.Keeper, sdk.Context, MockBankKeeper) {
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

	trackingKeeper := MockTrackingKeeper{
		GetBlockTrackingInfoFn: func(ctx sdk.Context, height int64) trackingtypes.BlockTracking {
			return trackingtypes.BlockTracking{
				Txs: nil,
			}
		},
	}
	bankKeeper := MockBankKeeper{
		BlockedAddrFn: func(addr sdk.AccAddress) bool { // everyaddress except distribution module address is blocked
			return addr.String() == authtypes.NewModuleAddress("distribution").String()
		},
		GetAllBalancesFn: func(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
			return sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 0))
		},
		SendCoinsFromModuleToAccountFn: func(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
			return nil
		},
	}
	authKeeper := MockAuthKeeper{
		GetModuleAccountFn: func(ctx context.Context, name string) sdk.ModuleAccountI {
			return MockModuleAccount{
				Address: "cosmos150j9auccvjdsquttx0qhawvux2m67fcxw49hy9",
			}
		},
	}
	k := keeper.NewKeeper(
		cdc,
		storeKey,
		nil,
		trackingKeeper,
		authKeeper,
		bankKeeper,
		"cosmos1a48wdtjn3egw7swhfkeshwdtjvs6hq9nlyrwut", // random addr for gov module
		log.NewTestLogger(tb),
	)
	ctx := sdk.NewContext(stateStore, tmproto.Header{
		Height: 1,
	}, false, log.NewNopLogger())
	err := k.Params.Set(ctx, types.DefaultParams())
	require.NoError(tb, err)
	return k, ctx, bankKeeper
}

type MockRewardsKeeper struct {
	ComputationalPriceOfGasFn func(ctx sdk.Context) sdk.DecCoin
	GetContractMetadataFn     func(ctx sdk.Context, contractAddr sdk.AccAddress) *types.ContractMetadata
}

func (k MockRewardsKeeper) ComputationalPriceOfGas(ctx sdk.Context) sdk.DecCoin {
	if k.ComputationalPriceOfGasFn == nil {
		panic("not supposed to be called!")
	}
	return k.ComputationalPriceOfGasFn(ctx)
}

func (k MockRewardsKeeper) GetContractMetadata(ctx sdk.Context, contractAddr sdk.AccAddress) *types.ContractMetadata {
	if k.GetContractMetadataFn == nil {
		panic("not supposed to be called!")
	}
	return k.GetContractMetadataFn(ctx, contractAddr)
}
