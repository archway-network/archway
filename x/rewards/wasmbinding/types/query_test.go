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
			name:  "OK: rewards query not used",
			query: Query{},
		},
		{
			name: "OK: Metadata",
			query: Query{
				Rewards: &RewardsQuery{
					Metadata: &ContractMetadataRequest{
						ContractAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
					},
				},
			},
		},
		{
			name: "OK: RewardsRecords",
			query: Query{
				Rewards: &RewardsQuery{
					RewardsRecords: &RewardsRecordsRequest{
						RewardsAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
					},
				},
			},
		},
		{
			name: "Fail: invalid Metadata",
			query: Query{
				Rewards: &RewardsQuery{
					Metadata: &ContractMetadataRequest{},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid RewardsRecords",
			query: Query{
				Rewards: &RewardsQuery{
					RewardsRecords: &RewardsRecordsRequest{},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: empty",
			query: Query{
				Rewards: &RewardsQuery{},
			},
			errExpected: true,
		},
		{
			name: "Fail: not one of",
			query: Query{
				Rewards: &RewardsQuery{
					Metadata: &ContractMetadataRequest{
						ContractAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
					},
					RewardsRecords: &RewardsRecordsRequest{
						RewardsAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
					},
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
