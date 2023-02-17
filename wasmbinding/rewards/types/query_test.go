package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContractMetadataRequestValidate(t *testing.T) {
	type testCase struct {
		name        string
		query       ContractMetadataRequest
		errExpected bool
	}

	testCases := []testCase{
		{
			name: "OK: Metadata",
			query: ContractMetadataRequest{
				ContractAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
			},
		},
		{
			name:        "Fail: invalid Metadata",
			query:       ContractMetadataRequest{},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.query.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestRewardsRecordsRequestValidate(t *testing.T) {
	type testCase struct {
		name        string
		query       RewardsRecordsRequest
		errExpected bool
	}

	testCases := []testCase{
		{
			name: "OK: RewardsRecords",
			query: RewardsRecordsRequest{
				RewardsAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
			},
		},
		{
			name:        "Fail: invalid RewardsRecords",
			query:       RewardsRecordsRequest{},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.query.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestContractFlatFeeRequestValidate(t *testing.T) {
	type testCase struct {
		name        string
		query       ContractFlatFeeRequest
		errExpected bool
	}

	testCases := []testCase{
		{
			name:        "Fail: Empty req",
			query:       ContractFlatFeeRequest{},
			errExpected: true,
		},
		{
			name: "Fail: Invalid req",
			query: ContractFlatFeeRequest{
				ContractAddress: "👻",
			},
			errExpected: true,
		},
		{
			name: "OK: Valid req",
			query: ContractFlatFeeRequest{
				ContractAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.query.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
