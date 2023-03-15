package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/spm/cosmoscmd"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"

	"github.com/archway-network/archway/app"
	"github.com/archway-network/archway/x/mint/keeper"
	"github.com/archway-network/archway/x/mint/types"
)

func SetupTestMintKeeper(t testing.TB) (keeper.Keeper, sdk.Context) {
	encoding := cosmoscmd.MakeEncodingConfig(app.ModuleBasics)
	appCodec := encoding.Marshaler
	cdc := encoding.Amino

	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	tStoreKey := sdk.NewTransientStoreKey("transient_test")

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(tStoreKey, sdk.StoreTypeTransient, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	marshaler := codec.NewProtoCodec(registry)

	paramsKeeper := paramskeeper.NewKeeper(appCodec, cdc, storeKey, tStoreKey)
	paramsKeeper.Subspace(types.ModuleName).WithKeyTable(types.ParamKeyTable())
	subspace, _ := paramsKeeper.GetSubspace(types.ModuleName)

	sk := MockStakingKeeper{
		BondedRatioFn: func(ctx sdk.Context) sdk.Dec {
			return sdk.MustNewDecFromStr("0.66")
		},
		BondDenomFn: func(ctx sdk.Context) string {
			return "stake"
		},
	}
	bk := MockBankKeeper{
		MintCoinsFn: func(ctx sdk.Context, name string, amt sdk.Coins) error {
			return nil
		},
		GetSupplyFn: func(ctx sdk.Context, denom string) sdk.Coin {
			return sdk.NewInt64Coin("stake", 50)
		},
		SendCoinsFromModuleToModuleFn: func(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error {
			return nil
		},
	}
	k := keeper.NewKeeper(marshaler, storeKey, subspace, bk, sk)
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	k.SetParams(ctx, types.DefaultParams())
	return k, ctx.WithBlockTime(time.Now())
}

type MockStakingKeeper struct {
	BondedRatioFn func(ctx sdk.Context) sdk.Dec
	BondDenomFn   func(ctx sdk.Context) string
}

func (k MockStakingKeeper) BondedRatio(ctx sdk.Context) sdk.Dec {
	if k.BondedRatioFn == nil {
		panic("not supposed to be called!")
	}
	return k.BondedRatioFn(ctx)
}

func (k MockStakingKeeper) BondDenom(ctx sdk.Context) string {
	if k.BondDenomFn == nil {
		panic("not supposed to be called!")
	}
	return k.BondDenomFn(ctx)
}

type MockBankKeeper struct {
	MintCoinsFn                   func(ctx sdk.Context, name string, amt sdk.Coins) error
	GetSupplyFn                   func(ctx sdk.Context, denom string) sdk.Coin
	SendCoinsFromModuleToModuleFn func(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
}

func (k MockBankKeeper) MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error {
	if k.MintCoinsFn == nil {
		panic("not supposed to be called!")
	}
	return k.MintCoinsFn(ctx, name, amt)
}

func (k MockBankKeeper) GetSupply(ctx sdk.Context, denom string) sdk.Coin {
	if k.GetSupplyFn == nil {
		panic("not supposed to be called!")
	}
	return k.GetSupplyFn(ctx, denom)
}

func (k MockBankKeeper) SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error {
	if k.SendCoinsFromModuleToModuleFn == nil {
		panic("not supposed to be called!")
	}
	return k.SendCoinsFromModuleToModuleFn(ctx, senderModule, recipientModule, amt)
}
