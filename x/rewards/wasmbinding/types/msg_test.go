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
			name: "OK: rewards sub-msg not used",
			msg:  Msg{},
		},
		{
			name: "OK: UpdateMetadata",
			msg: Msg{
				Rewards: &RewardsMsg{
					UpdateMetadata: &UpdateMetadataRequest{
						OwnerAddress:   "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
						RewardsAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
					},
				},
			},
		},
		{
			name: "OK: WithdrawRewards 1",
			msg: Msg{
				Rewards: &RewardsMsg{
					WithdrawRewards: &WithdrawRewardsRequest{
						RecordsLimit: 1,
					},
				},
			},
		},
		{
			name: "OK: WithdrawRewards 2",
			msg: Msg{
				Rewards: &RewardsMsg{
					WithdrawRewards: &WithdrawRewardsRequest{
						RecordIDs: []uint64{1},
					},
				},
			},
		},
		{
			name: "Fail: invalid UpdateMetadataRequest: no changes",
			msg: Msg{
				Rewards: &RewardsMsg{
					UpdateMetadata: &UpdateMetadataRequest{},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid UpdateMetadataRequest: invalid OwnerAddress",
			msg: Msg{
				Rewards: &RewardsMsg{
					UpdateMetadata: &UpdateMetadataRequest{
						OwnerAddress: "invalid",
					},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid UpdateMetadataRequest: invalid RewardsAddress",
			msg: Msg{
				Rewards: &RewardsMsg{
					UpdateMetadata: &UpdateMetadataRequest{
						RewardsAddress: "invalid",
					},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid WithdrawRewards: empty",
			msg: Msg{
				Rewards: &RewardsMsg{
					WithdrawRewards: &WithdrawRewardsRequest{},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid WithdrawRewards: one of failed",
			msg: Msg{
				Rewards: &RewardsMsg{
					WithdrawRewards: &WithdrawRewardsRequest{
						RecordsLimit: 1,
						RecordIDs:    []uint64{1},
					},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid WithdrawRewards: RecordIDs: invalid ID",
			msg: Msg{
				Rewards: &RewardsMsg{
					WithdrawRewards: &WithdrawRewardsRequest{
						RecordIDs: []uint64{1, 0},
					},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid WithdrawRewards: RecordIDs: duplicated IDs",
			msg: Msg{
				Rewards: &RewardsMsg{
					WithdrawRewards: &WithdrawRewardsRequest{
						RecordIDs: []uint64{1, 2, 1},
					},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: empty",
			msg: Msg{
				Rewards: &RewardsMsg{},
			},
			errExpected: true,
		},
		{
			name: "Fail: not one of",
			msg: Msg{
				Rewards: &RewardsMsg{
					UpdateMetadata: &UpdateMetadataRequest{
						OwnerAddress:   "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
						RewardsAddress: "cosmos1zj8lgj0zp06c8n4rreyzgu3tls9yhy4mm4vu8c",
					},
					WithdrawRewards: &WithdrawRewardsRequest{},
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
