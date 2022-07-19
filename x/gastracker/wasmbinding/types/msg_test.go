package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMsgValidate(t *testing.T) {
	type testCase struct {
		name        string
		msg         Msg
		errExpected bool
	}

	testCases := []testCase{
		{
			name: "OK",
			msg: Msg{
				UpdateMetadata: &UpdateMetadataRequest{
					DeveloperAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
				},
			},
		},
		{
			name: "Fail: invalid UpdateMetadataRequest: no changes",
			msg: Msg{
				UpdateMetadata: &UpdateMetadataRequest{},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid UpdateMetadataRequest: invalid DeveloperAddress",
			msg: Msg{
				UpdateMetadata: &UpdateMetadataRequest{
					DeveloperAddress: "invalid",
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid UpdateMetadataRequest: invalid RewardAddress",
			msg: Msg{
				UpdateMetadata: &UpdateMetadataRequest{
					RewardAddress: "invalid",
				},
			},
			errExpected: true,
		},
		{
			name:        "Fail: empty",
			msg:         Msg{},
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
