package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	trackingTypes "github.com/archway-network/archway/x/tracking/types"
)

func TestTrackingGenesisStateValidation(t *testing.T) {
	type testCase struct {
		name        string
		genesis     trackingTypes.GenesisState
		errExpected bool
	}

	contractAddrs := e2eTesting.GenContractAddresses(2)
	contractAddr1, contractAddr2 := contractAddrs[0], contractAddrs[1]

	testCases := []testCase{
		{
			name:    "OK: empty",
			genesis: trackingTypes.GenesisState{},
		},
		{
			name: "OK: non-empty",
			genesis: trackingTypes.GenesisState{
				TxInfoLastId: 1,
				TxInfos: []trackingTypes.TxInfo{
					{
						Id: 1,
					},
				},
				ContractOpInfoLastId: 2,
				ContractOpInfos: []trackingTypes.ContractOperationInfo{
					{
						Id:              1,
						TxId:            1,
						ContractAddress: contractAddr1.String(),
						OperationType:   trackingTypes.ContractOperation_CONTRACT_OPERATION_EXECUTION,
					},
					{
						Id:              2,
						TxId:            1,
						ContractAddress: contractAddr2.String(),
						OperationType:   trackingTypes.ContractOperation_CONTRACT_OPERATION_EXECUTION,
					},
				},
			},
		},
		{
			name: "Fail: invalid TxInfos",
			genesis: trackingTypes.GenesisState{
				TxInfos: []trackingTypes.TxInfo{
					{
						Id: 0,
					},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid ContractOperationInfos",
			genesis: trackingTypes.GenesisState{
				TxInfoLastId: 1,
				TxInfos: []trackingTypes.TxInfo{
					{
						Id: 1,
					},
				},
				ContractOpInfos: []trackingTypes.ContractOperationInfo{
					{
						Id: 0,
					},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: duplicated TxInfo",
			genesis: trackingTypes.GenesisState{
				TxInfoLastId: 1,
				TxInfos: []trackingTypes.TxInfo{
					{
						Id: 1,
					},
					{
						Id: 1,
					},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: duplicated ContractOperationInfos",
			genesis: trackingTypes.GenesisState{
				TxInfoLastId: 1,
				TxInfos: []trackingTypes.TxInfo{
					{
						Id: 1,
					},
				},
				ContractOpInfoLastId: 1,
				ContractOpInfos: []trackingTypes.ContractOperationInfo{
					{
						Id:              1,
						TxId:            1,
						ContractAddress: contractAddr1.String(),
						OperationType:   trackingTypes.ContractOperation_CONTRACT_OPERATION_EXECUTION,
					},
					{
						Id:              1,
						TxId:            1,
						ContractAddress: contractAddr2.String(),
						OperationType:   trackingTypes.ContractOperation_CONTRACT_OPERATION_EXECUTION,
					},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: unmatched TxID",
			genesis: trackingTypes.GenesisState{
				TxInfoLastId: 1,
				TxInfos: []trackingTypes.TxInfo{
					{
						Id: 1,
					},
				},
				ContractOpInfoLastId: 1,
				ContractOpInfos: []trackingTypes.ContractOperationInfo{
					{
						Id:              1,
						TxId:            2,
						ContractAddress: contractAddr1.String(),
						OperationType:   trackingTypes.ContractOperation_CONTRACT_OPERATION_EXECUTION,
					},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid lastTxID",
			genesis: trackingTypes.GenesisState{
				TxInfoLastId: 0,
				TxInfos: []trackingTypes.TxInfo{
					{
						Id: 1,
					},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid lastOpInfoID",
			genesis: trackingTypes.GenesisState{
				TxInfoLastId: 1,
				TxInfos: []trackingTypes.TxInfo{
					{
						Id: 1,
					},
				},
				ContractOpInfoLastId: 0,
				ContractOpInfos: []trackingTypes.ContractOperationInfo{
					{
						Id:              1,
						TxId:            1,
						ContractAddress: contractAddr1.String(),
						OperationType:   trackingTypes.ContractOperation_CONTRACT_OPERATION_EXECUTION,
					},
				},
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.genesis.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestTrackingTxInfoValidate(t *testing.T) {
	type testCase struct {
		name        string
		txInfo      trackingTypes.TxInfo
		errExpected bool
	}

	testCases := []testCase{
		{
			name: "OK",
			txInfo: trackingTypes.TxInfo{
				Id:       1,
				Height:   1,
				TotalGas: 100,
			},
		},
		{
			name: "Fail: invalid ID",
			txInfo: trackingTypes.TxInfo{
				Id: 0,
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.txInfo.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestTrackingContractOperationInfoValidate(t *testing.T) {
	type testCase struct {
		name        string
		opInfo      trackingTypes.ContractOperationInfo
		errExpected bool
	}

	contractAddr := e2eTesting.GenContractAddresses(1)[0]

	testCases := []testCase{
		{
			name: "OK",
			opInfo: trackingTypes.ContractOperationInfo{
				Id:              1,
				TxId:            1,
				ContractAddress: contractAddr.String(),
				OperationType:   trackingTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
				VmGas:           100,
				SdkGas:          50,
			},
		},
		{
			name: "Fail: invalid ID",
			opInfo: trackingTypes.ContractOperationInfo{
				Id:              0,
				TxId:            1,
				ContractAddress: contractAddr.String(),
				OperationType:   trackingTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid TxID",
			opInfo: trackingTypes.ContractOperationInfo{
				Id:              1,
				TxId:            0,
				ContractAddress: contractAddr.String(),
				OperationType:   trackingTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid ContractAddress",
			opInfo: trackingTypes.ContractOperationInfo{
				Id:              1,
				TxId:            1,
				ContractAddress: "invalid",
				OperationType:   trackingTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid OperationType",
			opInfo: trackingTypes.ContractOperationInfo{
				Id:              1,
				TxId:            1,
				ContractAddress: contractAddr.String(),
				OperationType:   100,
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.opInfo.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
