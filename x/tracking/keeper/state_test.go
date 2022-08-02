package keeper_test

import (
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/tracking/types"
)

// TestStates tests TxInfo and ContractOperationInfo state storages.
// Test append multiple objects for different blocks to make sure there are no namespace
// collisions (prefixed store keys) and state indexes work as expected.
// Final test stage is the cascade delete of all objects.
func (s *KeeperTestSuite) TestStates() {
	type testData struct {
		Tx  types.TxInfo
		Ops []types.ContractOperationInfo
	}

	chain := s.chain
	ctx, keeper := chain.GetContext(), chain.GetApp().TrackingKeeper

	// Fixtures
	startBlock := ctx.BlockHeight()

	testDataExpected := []testData{
		// Block 1, Tx 1: 3 ops
		{
			Tx: types.TxInfo{
				Id:       1,
				Height:   startBlock + 1,
				TotalGas: 450,
			},
			Ops: []types.ContractOperationInfo{
				{
					Id:              1,
					TxId:            1,
					ContractAddress: chain.GetAccount(0).Address.String(),
					OperationType:   types.ContractOperation_CONTRACT_OPERATION_EXECUTION,
					VmGas:           100, // here and below: converted to SDK gas
					SdkGas:          200,
				},
				{
					Id:              2,
					TxId:            1,
					ContractAddress: chain.GetAccount(1).Address.String(),
					OperationType:   types.ContractOperation_CONTRACT_OPERATION_QUERY,
					VmGas:           50,
					SdkGas:          100,
				},
			},
		},
		// Block 1, Tx 2: 3 ops (2 from the same contract)
		{
			Tx: types.TxInfo{
				Id:       2,
				Height:   startBlock + 1,
				TotalGas: 2600,
			},
			Ops: []types.ContractOperationInfo{
				{
					Id:              3,
					TxId:            2,
					ContractAddress: chain.GetAccount(2).Address.String(),
					OperationType:   types.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
					VmGas:           500,
					SdkGas:          1000,
				},
				{
					Id:              4,
					TxId:            2,
					ContractAddress: chain.GetAccount(3).Address.String(),
					OperationType:   types.ContractOperation_CONTRACT_OPERATION_EXECUTION,
					VmGas:           250,
					SdkGas:          300,
				},
				{
					Id:              5,
					TxId:            2,
					ContractAddress: chain.GetAccount(3).Address.String(),
					OperationType:   types.ContractOperation_CONTRACT_OPERATION_EXECUTION,
					VmGas:           250,
					SdkGas:          300,
				},
			},
		},
		// Block 2, Tx 1: 3 ops from 2 contracts (mixed)
		{
			Tx: types.TxInfo{
				Id:       3,
				Height:   startBlock + 2,
				TotalGas: 725,
			},
			Ops: []types.ContractOperationInfo{
				{
					Id:              6,
					TxId:            3,
					ContractAddress: chain.GetAccount(0).Address.String(),
					OperationType:   types.ContractOperation_CONTRACT_OPERATION_EXECUTION,
					VmGas:           50,
					SdkGas:          25,
				},
				{
					Id:              7,
					TxId:            3,
					ContractAddress: chain.GetAccount(1).Address.String(),
					OperationType:   types.ContractOperation_CONTRACT_OPERATION_IBC,
					VmGas:           100,
					SdkGas:          50,
				},
				{
					Id:              8,
					TxId:            3,
					ContractAddress: chain.GetAccount(0).Address.String(),
					OperationType:   types.ContractOperation_CONTRACT_OPERATION_REPLY,
					VmGas:           200,
					SdkGas:          300,
				},
			},
		},
		// Block 2, Tx 2: 2 ops from 2 contracts
		{
			Tx: types.TxInfo{
				Id:       4,
				Height:   startBlock + 2,
				TotalGas: 2100,
			},
			Ops: []types.ContractOperationInfo{
				{
					Id:              9,
					TxId:            4,
					ContractAddress: chain.GetAccount(0).Address.String(),
					OperationType:   types.ContractOperation_CONTRACT_OPERATION_MIGRATE,
					VmGas:           100,
					SdkGas:          500,
				},
				{
					Id:              10,
					TxId:            4,
					ContractAddress: chain.GetAccount(1).Address.String(),
					OperationType:   types.ContractOperation_CONTRACT_OPERATION_SUDO,
					VmGas:           500,
					SdkGas:          1000,
				},
			},
		},
	}

	// Upload fixtures
	block := int64(0)
	for _, data := range testDataExpected {
		// Switch to next block
		if data.Tx.Height != block {
			chain.NextBlock(0) // that updates TxInfo objs via EndBlocker
			ctx = chain.GetContext()
			block = ctx.BlockHeight()
		}

		// Start tracking a new Tx (emulate Ante handler) and check TxID sequence is correct
		keeper.TrackNewTx(ctx)
		s.Require().Equal(data.Tx.Id, keeper.GetState().TxInfoState(ctx).GetCurrentTxID())

		// Ingest contract operations
		records := make([]wasmTypes.ContractGasRecord, 0, len(data.Ops))
		for _, op := range data.Ops {
			records = append(
				records,
				wasmTypes.ContractGasRecord{
					OperationId:     testutils.RewardsContractOperationToWASM(op.OperationType),
					ContractAddress: op.ContractAddress,
					OriginalGas: wasmTypes.GasConsumptionInfo{
						VMGas:  keeper.WasmGasRegister.ToWasmVMGas(op.VmGas),
						SDKGas: op.SdkGas,
					},
				},
			)
		}
		s.Require().NoError(keeper.IngestGasRecord(ctx, records))
	}
	keeper.FinalizeBlockTxTracking(ctx)

	// Check non-existing records
	s.Run("Check non-existing state records", func() {
		_, txFound := keeper.GetState().TxInfoState(ctx).GetTxInfo(10)
		s.Assert().False(txFound)

		_, opFound := keeper.GetState().ContractOpInfoState(ctx).GetContractOpInfo(100)
		s.Assert().False(opFound)
	})

	// Check that the states are as expected
	s.Run("Check objects one by one", func() {
		opState := keeper.GetState().ContractOpInfoState(ctx)
		txState := keeper.GetState().TxInfoState(ctx)

		for _, data := range testDataExpected {
			// Check ContractOperations
			for _, op := range data.Ops {
				opInfo, found := opState.GetContractOpInfo(op.Id)
				s.Require().True(found, "ContractOpInfo (%d): not found", op.Id)
				s.Assert().Equal(op, opInfo, "ContractOpInfo (%d): wrong value", op.Id)
			}

			// Check TxInfo
			txInfo, found := txState.GetTxInfo(data.Tx.Id)
			s.Require().True(found, "TxInfo (%d): not found", data.Tx.Id)
			s.Assert().Equal(data.Tx, txInfo, "TxInfo (%d): wrong value", data.Tx.Id)
		}
	})

	// Check TxInfos search via block index
	s.Run("Check TxInfo block index", func() {
		txState := keeper.GetState().TxInfoState(ctx)

		// 1st block
		{
			height := testDataExpected[0].Tx.Height
			txInfosExpected := []types.TxInfo{
				testDataExpected[0].Tx,
				testDataExpected[1].Tx,
			}

			txInfosReceived := txState.GetTxInfosByBlock(height)
			s.Assert().ElementsMatch(txInfosExpected, txInfosReceived, "TxInfosByBlock (%d): wrong value", height)
		}

		// 2nd block
		{
			height := testDataExpected[2].Tx.Height
			txInfosExpected := []types.TxInfo{
				testDataExpected[2].Tx,
				testDataExpected[3].Tx,
			}

			txInfosReceived := txState.GetTxInfosByBlock(height)
			s.Assert().ElementsMatch(txInfosExpected, txInfosReceived, "TxInfosByBlock (%d): wrong value", height)
		}
	})

	// Check ContractOpInfos search via tx index
	s.Run("Check ContractOpInfo tx index", func() {
		opsState := keeper.GetState().ContractOpInfoState(ctx)

		for _, data := range testDataExpected {
			txID := data.Tx.Id
			opsExpected := data.Ops

			opsReceived := opsState.GetContractOpInfoByTxID(txID)
			s.Assert().ElementsMatch(opsExpected, opsReceived, "ContractOpInfoByTxID (%d): wrong value", txID)
		}
	})

	// Check records removal
	s.Run("Check records removal for the 1st block", func() {
		txState := keeper.GetState().TxInfoState(ctx)

		keeper.GetState().DeleteTxInfosCascade(ctx, startBlock+1)

		block1Txs := txState.GetTxInfosByBlock(startBlock + 1)
		s.Assert().Empty(block1Txs)

		block2Txs := txState.GetTxInfosByBlock(startBlock + 2)
		s.Assert().Len(block2Txs, 2)

		_, tx1Found := txState.GetTxInfo(testDataExpected[0].Tx.Id)
		s.Assert().False(tx1Found)

		_, tx2Found := txState.GetTxInfo(testDataExpected[1].Tx.Id)
		s.Assert().False(tx2Found)
	})

	s.Run("Check records removal for the 2nd block", func() {
		txState := keeper.GetState().TxInfoState(ctx)

		keeper.GetState().DeleteTxInfosCascade(ctx, startBlock+2)

		block2Txs := txState.GetTxInfosByBlock(startBlock + 2)
		s.Assert().Empty(block2Txs)

		_, tx3Found := txState.GetTxInfo(testDataExpected[2].Tx.Id)
		s.Assert().False(tx3Found)

		_, tx4Found := txState.GetTxInfo(testDataExpected[3].Tx.Id)
		s.Assert().False(tx4Found)
	})
}
