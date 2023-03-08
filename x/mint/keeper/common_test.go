package keeper_test

import (
	"testing"

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
	tStoreKey := sdk.NewTransientStoreKey(types.TStoreKey)

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

	k := keeper.NewKeeper(marshaler, storeKey, tStoreKey, subspace, nil, nil)
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	return k, ctx
}
