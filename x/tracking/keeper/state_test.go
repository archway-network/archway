package keeper_test

import (
	"testing"

	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/stretchr/testify/require"

	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/tracking"
	"github.com/archway-network/archway/x/tracking/types"
)

// TestStates tests TxInfo and ContractOperationInfo state storages.
// Test append multiple objects for different blocks to make sure there are no namespace
// collisions (prefixed store keys) and state indexes work as expected.
// Final test stage is the cascade delete of all objects.
func TestStates(t *testing.T) {
	type testData struct {
		Case string
		Tx   types.TxInfo
		Ops  []types.ContractOperationInfo
	}

	keeper, ctx := testutils.TrackingKeeper(t)

	// Fixtures
	startBlock := ctx.BlockHeight()

	testDataExpected := []testData{
		{
			Case: "Block 1, Tx 1: 3 ops",
			Tx: types.TxInfo{
				Id:       1,
				Height:   startBlock + 1,
				TotalGas: 450,
			},
			Ops: []types.ContractOperationInfo{
				{
					Id:              1,
					TxId:            1,
					ContractAddress: testutils.AccAddress().String(),
					OperationType:   types.ContractOperation_CONTRACT_OPERATION_EXECUTION,
					VmGas:           100, // here and below: converted to SDK gas
					SdkGas:          200,
				},
				{
					Id:              2,
					TxId:            1,
					ContractAddress: testutils.AccAddress().String(),
					OperationType:   types.ContractOperation_CONTRACT_OPERATION_QUERY,
					VmGas:           50,
					SdkGas:          100,
				},
			},
		},
		{
			Case: "Block 1, Tx 2: 3 ops (2 from the same contract)",
			Tx: types.TxInfo{
				Id:       2,
				Height:   startBlock + 1,
				TotalGas: 2600,
			},
			Ops: []types.ContractOperationInfo{
				{
					Id:              3,
					TxId:            2,
					ContractAddress: testutils.AccAddress().String(),
					OperationType:   types.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
					VmGas:           500,
					SdkGas:          1000,
				},
				{
					Id:              4,
					TxId:            2,
					ContractAddress: testutils.AccAddress().String(),
					OperationType:   types.ContractOperation_CONTRACT_OPERATION_EXECUTION,
					VmGas:           250,
					SdkGas:          300,
				},
				{
					Id:              5,
					TxId:            2,
					ContractAddress: testutils.AccAddress().String(),
					OperationType:   types.ContractOperation_CONTRACT_OPERATION_EXECUTION,
					VmGas:           250,
					SdkGas:          300,
				},
			},
		},
		{
			Case: "Block 2, Tx 1: 3 ops from 2 contracts (mixed)",
			Tx: types.TxInfo{
				Id:       3,
				Height:   startBlock + 2,
				TotalGas: 725,
			},
			Ops: []types.ContractOperationInfo{
				{
					Id:              6,
					TxId:            3,
					ContractAddress: testutils.AccAddress().String(),
					OperationType:   types.ContractOperation_CONTRACT_OPERATION_EXECUTION,
					VmGas:           50,
					SdkGas:          25,
				},
				{
					Id:              7,
					TxId:            3,
					ContractAddress: testutils.AccAddress().String(),
					OperationType:   types.ContractOperation_CONTRACT_OPERATION_IBC,
					VmGas:           100,
					SdkGas:          50,
				},
				{
					Id:              8,
					TxId:            3,
					ContractAddress: testutils.AccAddress().String(),
					OperationType:   types.ContractOperation_CONTRACT_OPERATION_REPLY,
					VmGas:           200,
					SdkGas:          300,
				},
			},
		},
		{
			Case: "Block 2, Tx 2: 2 ops from 2 contracts",
			Tx: types.TxInfo{
				Id:       4,
				Height:   startBlock + 2,
				TotalGas: 2100,
			},
			Ops: []types.ContractOperationInfo{
				{
					Id:              9,
					TxId:            4,
					ContractAddress: testutils.AccAddress().String(),
					OperationType:   types.ContractOperation_CONTRACT_OPERATION_MIGRATE,
					VmGas:           100,
					SdkGas:          500,
				},
				{
					Id:              10,
					TxId:            4,
					ContractAddress: testutils.AccAddress().String(),
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
			_, err := tracking.EndBlocker(ctx, keeper) // that updates TxInfo objs via EndBlocker
			require.NoError(t, err)
			ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

			block = ctx.BlockHeight()
		}

		// Start tracking a new Tx (emulate Ante handler) and check TxID sequence is correct
		keeper.TrackNewTx(ctx)
		require.Equal(t, data.Tx.Id, keeper.GetState().TxInfoState(ctx).GetCurrentTxID())

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
		require.NoError(t, keeper.IngestGasRecord(ctx, records))
	}
	keeper.FinalizeBlockTxTracking(ctx)

	// Check non-existing records
	t.Run("Check non-existing state records", func(t *testing.T) {
		_, txFound := keeper.GetState().TxInfoState(ctx).GetTxInfo(10)
		require.False(t, txFound)

		_, opFound := keeper.GetState().ContractOpInfoState(ctx).GetContractOpInfo(100)
		require.False(t, opFound)
	})

	// Check that the states are as expected
	t.Run("Check objects one by one", func(t *testing.T) {
		opState := keeper.GetState().ContractOpInfoState(ctx)
		txState := keeper.GetState().TxInfoState(ctx)

		for _, data := range testDataExpected {
			// Check ContractOperations
			for _, op := range data.Ops {
				opInfo, found := opState.GetContractOpInfo(op.Id)
				require.True(t, found, "ContractOpInfo (%d): not found", op.Id)
				require.Equal(t, op, opInfo, "ContractOpInfo (%d): wrong value", op.Id)
			}

			// Check TxInfo
			txInfo, found := txState.GetTxInfo(data.Tx.Id)
			require.True(t, found, "TxInfo (%d): not found", data.Tx.Id)
			require.Equal(t, data.Tx, txInfo, "TxInfo (%d): wrong value", data.Tx.Id)
		}
	})

	// Check TxInfos search via block index
	t.Run("Check TxInfo block index", func(t *testing.T) {
		txState := keeper.GetState().TxInfoState(ctx)

		// 1st block
		{
			height := testDataExpected[0].Tx.Height
			txInfosExpected := []types.TxInfo{
				testDataExpected[0].Tx,
				testDataExpected[1].Tx,
			}

			txInfosReceived := txState.GetTxInfosByBlock(height)
			require.ElementsMatch(t, txInfosExpected, txInfosReceived, "TxInfosByBlock (%d): wrong value", height)
		}

		// 2nd block
		{
			height := testDataExpected[2].Tx.Height
			txInfosExpected := []types.TxInfo{
				testDataExpected[2].Tx,
				testDataExpected[3].Tx,
			}

			txInfosReceived := txState.GetTxInfosByBlock(height)
			require.ElementsMatch(t, txInfosExpected, txInfosReceived, "TxInfosByBlock (%d): wrong value", height)
		}
	})

	// Check ContractOpInfos search via tx index
	t.Run("Check ContractOpInfo tx index", func(t *testing.T) {
		opsState := keeper.GetState().ContractOpInfoState(ctx)

		for _, data := range testDataExpected {
			txID := data.Tx.Id
			opsExpected := data.Ops

			opsReceived := opsState.GetContractOpInfoByTxID(txID)
			require.ElementsMatch(t, opsExpected, opsReceived, "ContractOpInfoByTxID (%d): wrong value", txID)
		}
	})

	// Check records removal
	t.Run("Check records removal for the 1st block", func(t *testing.T) {
		txState := keeper.GetState().TxInfoState(ctx)

		keeper.GetState().DeleteTxInfosCascade(ctx, startBlock+1)

		block1Txs := txState.GetTxInfosByBlock(startBlock + 1)
		require.Empty(t, block1Txs)

		block2Txs := txState.GetTxInfosByBlock(startBlock + 2)
		require.Len(t, block2Txs, 2)

		_, tx1Found := txState.GetTxInfo(testDataExpected[0].Tx.Id)
		require.False(t, tx1Found)

		_, tx2Found := txState.GetTxInfo(testDataExpected[1].Tx.Id)
		require.False(t, tx2Found)
	})

	t.Run("Check records removal for the 2nd block", func(t *testing.T) {
		txState := keeper.GetState().TxInfoState(ctx)

		keeper.GetState().DeleteTxInfosCascade(ctx, startBlock+2)

		block2Txs := txState.GetTxInfosByBlock(startBlock + 2)
		require.Empty(t, block2Txs)

		_, tx3Found := txState.GetTxInfo(testDataExpected[2].Tx.Id)
		require.False(t, tx3Found)

		_, tx4Found := txState.GetTxInfo(testDataExpected[3].Tx.Id)
		require.False(t, tx4Found)
	})
}
