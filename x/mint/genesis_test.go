package mint_test

import (
	"testing"
	"time"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/mint"
	"github.com/archway-network/archway/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

// TestGenesisImportExport check genesis import/export.
func (s *KeeperTestSuite) TestGenesisImportExport() {
	currentTime := time.Now()
	ctx, k := s.chain.GetContext().WithBlockTime(currentTime), s.chain.GetApp().MintKeeper
	inflation := sdk.MustNewDecFromStr("0.3")

	s.Run("Panic: invalid inflation info. should panic", func() {
		initParams := types.DefaultParams()
		genesisState := types.GenesisState{
			Params: initParams,
			LastBlockInfo: types.LastBlockInfo{
				Inflation: sdk.MustNewDecFromStr("8"),
				Time:      &currentTime,
			},
		}
		s.Panics(func() { mint.InitGenesis(ctx, k, genesisState) })
	})

	s.Run("OK: check state matches init input. LastBlockInfo missing", func() {
		initParams := types.DefaultParams()
		initParams.MinInflation = sdk.MustNewDecFromStr("1")
		genesisState := types.GenesisState{
			Params: initParams,
		}

		mint.InitGenesis(ctx, k, genesisState)

		params := k.GetParams(ctx)
		lbi, found := k.GetLastBlockInfo(ctx)
		s.Require().EqualValues(initParams, params)
		s.Require().True(found)
		s.Require().EqualValues(params.MinInflation, lbi.Inflation)
		s.Require().EqualValues(currentTime.UTC(), lbi.Time.UTC())
	})

	s.Run("OK: check state matches init input", func() {
		initParams := types.DefaultParams()
		initParams.InflationChange = sdk.MustNewDecFromStr("0.123")
		genesisState := types.GenesisState{
			Params: initParams,
			LastBlockInfo: types.LastBlockInfo{
				Inflation: inflation,
				Time:      &currentTime,
			},
		}

		mint.InitGenesis(ctx, k, genesisState)

		params := k.GetParams(ctx)
		lbi, found := k.GetLastBlockInfo(ctx)
		s.Require().EqualValues(initParams, params)
		s.Require().True(found)
		s.Require().EqualValues(inflation, lbi.Inflation)
		s.Require().EqualValues(currentTime.UTC(), lbi.Time.UTC())
	})

	s.Run("OK: check export matches what we init", func() {
		k.SetParams(ctx, types.DefaultParams())
		genesisState := mint.ExportGenesis(ctx, k)
		s.Require().NotNil(genesisState)
		s.Require().EqualValues(types.DefaultParams(), genesisState.Params)
		s.Require().EqualValues(inflation, genesisState.LastBlockInfo.Inflation)
		s.Require().EqualValues(currentTime.UTC(), genesisState.LastBlockInfo.Time.UTC())
	})
}

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
