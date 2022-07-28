package keeper_test

import (
	"math/rand"

	"github.com/archway-network/archway/x/tracking/types"
)

func (s *KeeperTestSuite) TestGenesisExport() {
	rand.Seed(86)
	chain := s.chain
	ctx, keeper := chain.GetContext(), chain.GetApp().TrackingKeeper
	operationsToExecute := 100

	for i := 0; i < operationsToExecute; i++ {
		keeper.TrackNewTx(ctx)
		keeper.TrackNewContractOperation(ctx, chain.GetAccount(rand.Intn(5)).Address, types.ContractOperation(rand.Int31n(8)-1), 1, 1)
	}
	p := keeper.GetParams(ctx)
	p.GasTrackingEnabled = false
	keeper.SetParams(ctx, p)

	genesis := keeper.ExportGenesis(ctx)
	s.Require().Nil(genesis.Validate())
	s.Require().False(genesis.Params.GasTrackingEnabled)
	s.Require().Equal(operationsToExecute, len(genesis.ContractOpInfos))
	s.Require().Equal(operationsToExecute, len(genesis.TxInfos))
}

func (s *KeeperTestSuite) TestGenesisImport() {
	chain := s.chain
	ctx, keeper := chain.GetContext(), chain.GetApp().TrackingKeeper
	genesis := types.DefaultGenesisState()
	genesis.Params.GasTrackingEnabled = false
	operationsToExecute := 100

	// Ids must be greater than 0
	for i := 1; i <= operationsToExecute; i++ {
		txInfo := types.TxInfo{uint64(i), 0, 2}
		contractOperation := types.ContractOperationInfo{
			txInfo.Id,
			txInfo.Id,
			chain.GetAccount(rand.Intn(5)).Address.String(),
			types.ContractOperation(rand.Int31n(8) - 1),
			1,
			1,
		}
		genesis.TxInfos = append(genesis.TxInfos, txInfo)
		genesis.ContractOpInfos = append(genesis.ContractOpInfos, contractOperation)
	}

	s.Require().Nil(genesis.Validate())
	keeper.InitGenesis(ctx, genesis)
	newGenesis := keeper.ExportGenesis(ctx)
	s.Require().Equal(genesis, newGenesis)
}
