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
