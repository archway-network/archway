package types

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/archway-network/archway/wasmbinding/rewards/types"
)

func TestMsgValidate(t *testing.T) {
	type testCase struct {
		name        string
		msg         Msg
		errExpected bool
	}

	testCases := []testCase{
		{
			name: "OK: Rewards",
			msg: Msg{
				UpdateContractMetadata: &types.UpdateContractMetadataRequest{
					OwnerAddress:   "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
					RewardsAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
				},
			},
		},
		{
			name: "Fail: not one of",
			msg: Msg{
				UpdateContractMetadata: &types.UpdateContractMetadataRequest{
					OwnerAddress:   "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
					RewardsAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
				},
				WithdrawRewards: &types.WithdrawRewardsRequest{},
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
