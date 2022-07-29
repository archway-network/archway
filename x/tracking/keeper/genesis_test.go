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

	genesis := keeper.ExportGenesis(ctx)
	s.Require().NoError(genesis.Validate())
	s.Assert().Equal(operationsToExecute, len(genesis.TxInfos))
	s.Assert().Equal(operationsToExecute, len(genesis.ContractOpInfos))
}

func (s *KeeperTestSuite) TestGenesisImport() {
	chain := s.chain
	ctx, keeper := chain.GetContext(), chain.GetApp().TrackingKeeper
	genesisExpected := types.DefaultGenesisState()
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
		genesisExpected.TxInfos = append(genesisExpected.TxInfos, txInfo)
		genesisExpected.ContractOpInfos = append(genesisExpected.ContractOpInfos, contractOperation)
	}
	s.Require().NoError(genesisExpected.Validate())

	keeper.InitGenesis(ctx, genesisExpected)
	genesisReceived := keeper.ExportGenesis(ctx)
	s.Assert().Equal(genesisExpected, genesisReceived)
}
