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
			name: "OK",
			query: Query{
				Metadata: &ContractMetadataRequest{
					ContractAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
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
			name:        "Fail: empty",
			query:       Query{},
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
