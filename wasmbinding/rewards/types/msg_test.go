package types

import (
	"testing"

	wasmVmTypes "github.com/CosmWasm/wasmvm/types"
	"github.com/stretchr/testify/assert"

	"github.com/archway-network/archway/pkg"
)

func TestUpdateContractMetadataRequestValidate(t *testing.T) {
	type testCase struct {
		name        string
		msg         UpdateContractMetadataRequest
		errExpected bool
	}

	testCases := []testCase{
		{
			name: "OK: UpdateMetadata",
			msg: UpdateContractMetadataRequest{
				ContractAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
				OwnerAddress:    "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
				RewardsAddress:  "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
			},
		},
		{
			name:        "Fail: invalid UpdateMetadataRequest: no changes",
			msg:         UpdateContractMetadataRequest{},
			errExpected: true,
		},
		{
			name: "Fail: invalid UpdateMetadataRequest: invalid ContractAddress",
			msg: UpdateContractMetadataRequest{
				ContractAddress: "invalid",
				OwnerAddress:    "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid UpdateMetadataRequest: invalid OwnerAddress",
			msg: UpdateContractMetadataRequest{
				OwnerAddress: "invalid",
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid UpdateMetadataRequest: invalid RewardsAddress",
			msg: UpdateContractMetadataRequest{
				RewardsAddress: "invalid",
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestUpdateContractMetadataRequestMustGetContractAddressOk(t *testing.T) {
	type testCase struct {
		name         string
		msg          UpdateContractMetadataRequest
		addrExpected bool
	}

	testCases := []testCase{
		{
			name: "UpdateMetadata has Contract Address",
			msg: UpdateContractMetadataRequest{
				ContractAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
			},
			addrExpected: true,
		},
		{
			name:         "UpdateMetadataRequest does not have Contract Address",
			msg:          UpdateContractMetadataRequest{},
			addrExpected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			addr, isSet := tc.msg.MustGetContractAddressOk()
			if tc.addrExpected {
				assert.NotEmpty(t, addr)
				assert.True(t, isSet)
				return
			}
			assert.Empty(t, addr)
			assert.False(t, isSet)
		})
	}
}

func TestWithdrawRewardsRequestValidate(t *testing.T) {
	type testCase struct {
		name        string
		msg         WithdrawRewardsRequest
		errExpected bool
	}

	testCases := []testCase{
		{
			name: "OK: WithdrawRewards 1",
			msg: WithdrawRewardsRequest{
				RecordsLimit: pkg.Uint64Ptr(1),
			},
		},
		{
			name: "OK: WithdrawRewards 2",
			msg: WithdrawRewardsRequest{
				RecordIDs: []uint64{1},
			},
		},
		{
			name: "OK: WithdrawRewards 3",
			msg: WithdrawRewardsRequest{
				RecordsLimit: pkg.Uint64Ptr(0),
			},
		},
		{
			name:        "Fail: invalid WithdrawRewards: empty",
			msg:         WithdrawRewardsRequest{},
			errExpected: true,
		},
		{
			name: "Fail: invalid WithdrawRewards: one of failed",
			msg: WithdrawRewardsRequest{
				RecordsLimit: pkg.Uint64Ptr(1),
				RecordIDs:    []uint64{1},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid WithdrawRewards: RecordIDs: invalid ID",
			msg: WithdrawRewardsRequest{
				RecordIDs: []uint64{1, 0},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid WithdrawRewards: RecordIDs: duplicated IDs",
			msg: WithdrawRewardsRequest{
				RecordIDs: []uint64{1, 2, 1},
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestSetFlatFeeRequestValidate(t *testing.T) {
	type testCase struct {
		name        string
		msg         SetFlatFeeRequest
		errExpected bool
	}

	testCases := []testCase{
		{
			name: "OK: SetFlatFeeRequest",
			msg: SetFlatFeeRequest{
				ContractAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
				FlatFeeAmount: wasmVmTypes.Coin{
					Denom:  "test",
					Amount: "10",
				},
			},
		},
		{
			name:        "Fail: invalid SetFlatFeeRequest: no changes",
			msg:         SetFlatFeeRequest{},
			errExpected: true,
		},
		{
			name: "Fail: invalid SetFlatFeeRequest: invalid contractAddress",
			msg: SetFlatFeeRequest{
				ContractAddress: "👻",
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid SetFlatFeeRequest: invalid fee",
			msg: SetFlatFeeRequest{
				ContractAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
				FlatFeeAmount: wasmVmTypes.Coin{
					Denom:  "test",
					Amount: "👻",
				},
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
