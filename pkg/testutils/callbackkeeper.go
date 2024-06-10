package testutils

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	"github.com/archway-network/archway/x/callback/keeper"
	"github.com/archway-network/archway/x/callback/types"
	rewardstypes "github.com/archway-network/archway/x/rewards/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func CallbackKeeper(tb testing.TB) (keeper.Keeper, sdk.Context) {
	tb.Helper()
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey("m_callback")
	tStoreKey := storetypes.NewTransientStoreKey("t_callback")

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
			if amt[0].Amount.LTE(math.NewInt(3500000000)) {
				return nil
			}
			return sdkerrors.ErrInsufficientFunds
		},
		BlockedAddrFn: func(addr sdk.AccAddress) bool {
			return false
		},
		SendCoinsFromModuleToAccountFn: func(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
			return nil
		},
		SendCoinsFromModuleToModuleFn: func(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error {
			return nil
		},
	}
	rewardsKeeper := MockRewardsKeeper{
		ComputationalPriceOfGasFn: func(ctx sdk.Context) sdk.DecCoin {
			return sdk.NewInt64DecCoin(sdk.DefaultBondDenom, 1)
		},
		GetContractMetadataFn: func(ctx sdk.Context, contractAddr sdk.AccAddress) *rewardstypes.ContractMetadata {
			return &rewardstypes.ContractMetadata{}
		},
	}
	k := keeper.NewKeeper(
		cdc,
		storeKey,
		nil,
		rewardsKeeper,
		bankKeeper,
		"cosmos1a48wdtjn3egw7swhfkeshwdtjvs6hq9nlyrwut", // random addr for gov module
		log.NewTestLogger(tb),
	)
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	params := types.DefaultParams()
	_ = k.SetParams(ctx, params)

	return k, ctx
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
