package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

func TestMsgSetContractMetadataValidateBasic(t *testing.T) {
	type testCase struct {
		name        string
		msg         rewardsTypes.MsgSetContractMetadata
		errExpected bool
	}

	accAddrs, _ := e2eTesting.GenAccounts(1)
	accAddr := accAddrs[0]

	contractAddr := e2eTesting.GenContractAddresses(1)[0]

	testCases := []testCase{
		{
			name: "OK",
			msg: rewardsTypes.MsgSetContractMetadata{
				SenderAddress: accAddr.String(),
				Metadata: rewardsTypes.ContractMetadata{
					ContractAddress: contractAddr.String(),
				},
			},
		},
		{
			name: "Fail: invalid SenderAddress",
			msg: rewardsTypes.MsgSetContractMetadata{
				SenderAddress: "invalid",
				Metadata: rewardsTypes.ContractMetadata{
					ContractAddress: contractAddr.String(),
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid Metadata",
			msg: rewardsTypes.MsgSetContractMetadata{
				SenderAddress: accAddr.String(),
				Metadata: rewardsTypes.ContractMetadata{
					ContractAddress: "invalid",
				},
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestMsgWithdrawRewardsValidateBasic(t *testing.T) {
	type testCase struct {
		name        string
		msg         rewardsTypes.MsgWithdrawRewards
		errExpected bool
	}

	accAddrs, _ := e2eTesting.GenAccounts(1)
	accAddr := accAddrs[0]

	testCases := []testCase{
		{
			name: "OK 1",
			msg: rewardsTypes.MsgWithdrawRewards{
				RewardsAddress: accAddr.String(),
				Mode: &rewardsTypes.MsgWithdrawRewards_RecordsLimit_{
					RecordsLimit: &rewardsTypes.MsgWithdrawRewards_RecordsLimit{
						Limit: 1,
					},
				},
			},
		},
		{
			name: "OK 2",
			msg: rewardsTypes.MsgWithdrawRewards{
				RewardsAddress: accAddr.String(),
				Mode: &rewardsTypes.MsgWithdrawRewards_RecordIds{
					RecordIds: &rewardsTypes.MsgWithdrawRewards_RecordIDs{
						Ids: []uint64{1},
					},
				},
			},
		},
		{
			name: "Fail: invalid RewardsAddress",
			msg: rewardsTypes.MsgWithdrawRewards{
				RewardsAddress: "invalid",
			},
			errExpected: true,
		},
		{
			name: "Fail: no mode set",
			msg: rewardsTypes.MsgWithdrawRewards{
				RewardsAddress: accAddr.String(),
			},
			errExpected: true,
		},
		{
			name: "Fail: RecordsLimit: empty object",
			msg: rewardsTypes.MsgWithdrawRewards{
				RewardsAddress: accAddr.String(),
				Mode:           (*rewardsTypes.MsgWithdrawRewards_RecordsLimit_)(nil),
			},
			errExpected: true,
		},
		{
			name: "Fail: RecordsLimit: empty request",
			msg: rewardsTypes.MsgWithdrawRewards{
				RewardsAddress: accAddr.String(),
				Mode:           &rewardsTypes.MsgWithdrawRewards_RecordsLimit_{},
			},
			errExpected: true,
		},
		{
			name: "Fail: RecordsLimit: empty object",
			msg: rewardsTypes.MsgWithdrawRewards{
				RewardsAddress: accAddr.String(),
				Mode:           (*rewardsTypes.MsgWithdrawRewards_RecordIds)(nil),
			},
			errExpected: true,
		},
		{
			name: "Fail: RecordsLimit: empty request",
			msg: rewardsTypes.MsgWithdrawRewards{
				RewardsAddress: accAddr.String(),
				Mode:           &rewardsTypes.MsgWithdrawRewards_RecordIds{},
			},
			errExpected: true,
		},
		{
			name: "Fail: RecordsLimit: empty IDs",
			msg: rewardsTypes.MsgWithdrawRewards{
				RewardsAddress: accAddr.String(),
				Mode: &rewardsTypes.MsgWithdrawRewards_RecordIds{
					RecordIds: &rewardsTypes.MsgWithdrawRewards_RecordIDs{},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: RecordsLimit: invalid ID",
			msg: rewardsTypes.MsgWithdrawRewards{
				RewardsAddress: accAddr.String(),
				Mode: &rewardsTypes.MsgWithdrawRewards_RecordIds{
					RecordIds: &rewardsTypes.MsgWithdrawRewards_RecordIDs{
						Ids: []uint64{1, 0},
					},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: RecordsLimit: duplicated ID",
			msg: rewardsTypes.MsgWithdrawRewards{
				RewardsAddress: accAddr.String(),
				Mode: &rewardsTypes.MsgWithdrawRewards_RecordIds{
					RecordIds: &rewardsTypes.MsgWithdrawRewards_RecordIDs{
						Ids: []uint64{1, 2, 2, 3},
					},
				},
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestMsgSetFlatFeeValidateBasic(t *testing.T) {
	type testCase struct {
		name        string
		msg         rewardsTypes.MsgSetFlatFee
		errExpected bool
	}

	accAddrs, _ := e2eTesting.GenAccounts(1)
	accAddr := accAddrs[0]

	contractAddr := e2eTesting.GenContractAddresses(1)[0]

	testCases := []testCase{
		{
			name: "OK",
			msg: rewardsTypes.MsgSetFlatFee{
				SenderAddress:   accAddr.String(),
				ContractAddress: contractAddr.String(),
			},
		},
		{
			name: "Fail: invalid SenderAddress",
			msg: rewardsTypes.MsgSetFlatFee{
				SenderAddress:   "👻",
				ContractAddress: contractAddr.String(),
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid Contract Address",
			msg: rewardsTypes.MsgSetFlatFee{
				SenderAddress:   accAddr.String(),
				ContractAddress: "👻",
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
