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
			name: "OK 1",
			msg: Msg{
				UpdateMetadata: &UpdateMetadataRequest{
					OwnerAddress:   "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
					RewardsAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
				},
			},
		},
		{
			name: "OK 2",
			msg: Msg{
				WithdrawRewards: &WithdrawRewardsRequest{},
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
			name: "Fail: invalid UpdateMetadataRequest: invalid OwnerAddress",
			msg: Msg{
				UpdateMetadata: &UpdateMetadataRequest{
					OwnerAddress: "invalid",
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid UpdateMetadataRequest: invalid RewardsAddress",
			msg: Msg{
				UpdateMetadata: &UpdateMetadataRequest{
					RewardsAddress: "invalid",
				},
			},
			errExpected: true,
		},
		{
			name:        "Fail: empty",
			msg:         Msg{},
			errExpected: true,
		},
		{
			name: "Fail: not one of",
			msg: Msg{
				UpdateMetadata: &UpdateMetadataRequest{
					OwnerAddress:   "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
					RewardsAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
				},
				WithdrawRewards: &WithdrawRewardsRequest{},
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
