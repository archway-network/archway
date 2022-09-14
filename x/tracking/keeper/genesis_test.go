package keeper_test

import (
	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/tracking/types"
)

// TestGenesisImportExport check genesis import/export.
// Test updates the initial state with new txs and checks that they were merged.
func (s *KeeperTestSuite) TestGenesisImportExport() {
	ctx, keeper := s.chain.GetContext(), s.chain.GetApp().TrackingKeeper

	contractAddrs := e2eTesting.GenContractAddresses(2)

	var genesisStateInitial types.GenesisState
	s.Run("Check export of the initial genesis", func() {
		genesisState := keeper.ExportGenesis(ctx)
		s.Require().NotNil(genesisState)

		s.Assert().Empty(genesisState.TxInfoLastId)
		s.Assert().Empty(genesisState.TxInfos)
		s.Assert().Empty(genesisState.ContractOpInfos)

		genesisStateInitial = *genesisState
	})

	newTxInfos := []types.TxInfo{
		{
			Id:       110,
			Height:   100,
			TotalGas: 1000,
		},
		{
			Id:       210,
			Height:   200,
			TotalGas: 2000,
		},
	}

	newContractOpInfos := []types.ContractOperationInfo{
		{
			Id:              1,
			TxId:            110,
			ContractAddress: contractAddrs[0].String(),
			OperationType:   testutils.WASMContractOperationToRewards(testutils.GetRandomContractOperationType()),
			VmGas:           150,
			SdkGas:          250,
		},
		{
			Id:              2,
			TxId:            210,
			ContractAddress: contractAddrs[1].String(),
			OperationType:   testutils.WASMContractOperationToRewards(testutils.GetRandomContractOperationType()),
			VmGas:           350,
			SdkGas:          450,
		},
	}

	genesisStateImported := types.NewGenesisState(
		newTxInfos[len(newTxInfos)-1].Id,
		newTxInfos,
		newContractOpInfos[len(newContractOpInfos)-1].Id,
		newContractOpInfos,
	)
	s.Run("Check import of an updated genesis", func() {
		keeper.InitGenesis(ctx, genesisStateImported)

		genesisStateExpected := types.GenesisState{
			TxInfoLastId:         newTxInfos[len(newTxInfos)-1].Id,
			TxInfos:              append(genesisStateInitial.TxInfos, newTxInfos...),
			ContractOpInfoLastId: newContractOpInfos[len(newContractOpInfos)-1].Id,
			ContractOpInfos:      append(genesisStateInitial.ContractOpInfos, newContractOpInfos...),
		}

		genesisStateReceived := keeper.ExportGenesis(ctx)
		s.Require().NotNil(genesisStateReceived)
		s.Assert().Equal(genesisStateExpected.TxInfoLastId, genesisStateReceived.TxInfoLastId)
		s.Assert().ElementsMatch(genesisStateExpected.TxInfos, genesisStateReceived.TxInfos)
		s.Assert().Equal(genesisStateExpected.ContractOpInfoLastId, genesisStateReceived.ContractOpInfoLastId)
		s.Assert().ElementsMatch(genesisStateExpected.ContractOpInfos, genesisStateReceived.ContractOpInfos)
	})
}
