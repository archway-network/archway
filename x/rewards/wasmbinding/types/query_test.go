package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryValidate(t *testing.T) {
	type testCase struct {
		name        string
		query       Query
		errExpected bool
	}

	testCases := []testCase{
		{
			name: "OK 1",
			query: Query{
				Metadata: &ContractMetadataRequest{
					ContractAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
				},
			},
		},
		{
			name: "OK 2",
			query: Query{
				CurrentRewards: &CurrentRewardsRequest{
					RewardsAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
				},
			},
		},
		{
			name: "Fail: invalid Metadata",
			query: Query{
				Metadata: &ContractMetadataRequest{},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid CurrentRewards",
			query: Query{
				CurrentRewards: &CurrentRewardsRequest{},
			},
			errExpected: true,
		},
		{
			name:        "Fail: empty",
			query:       Query{},
			errExpected: true,
		},
		{
			name: "Fail: not one of",
			query: Query{
				Metadata: &ContractMetadataRequest{
					ContractAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
				},
				CurrentRewards: &CurrentRewardsRequest{
					RewardsAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
				},
			},
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
