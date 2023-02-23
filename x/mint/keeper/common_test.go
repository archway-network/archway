package keeper_test

import (
	"testing"

	"github.com/archway-network/archway/app"
	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/mint/keeper"
	"github.com/archway-network/archway/x/mint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/spm/cosmoscmd"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
)

type KeeperTestSuite struct {
	suite.Suite
	chain *e2eTesting.TestChain
}

func (s *KeeperTestSuite) SetupTest() {
	s.chain = e2eTesting.NewTestChain(s.T(), 1)
}

func TestMintKeeper(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func SetupTestMintKeeper(t testing.TB) (keeper.Keeper, sdk.Context) {
	encoding := cosmoscmd.MakeEncodingConfig(app.ModuleBasics)
	appCodec := encoding.Marshaler
	cdc := encoding.Amino

	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	tStoreKey := sdk.NewTransientStoreKey(types.StoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	marshaler := codec.NewProtoCodec(registry)

	paramsKeeper := paramskeeper.NewKeeper(appCodec, cdc, storeKey, tStoreKey)
	paramsKeeper.Subspace(types.ModuleName)
	subspace, _ := paramsKeeper.GetSubspace(types.ModuleName)

	k := keeper.NewKeeper(marshaler, storeKey, subspace)
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	return k, ctx

}
