package keeper_test

import (
	"fmt"
	"github.com/archway-network/archway/pkg/testutils"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/rewards/keeper"
	rewardstypes "github.com/archway-network/archway/x/rewards/types"
)

func (s *KeeperTestSuite) TestMsgServer_SetContractMetadata() {
	testCases := []struct {
		testCase    string
		prepare     func() *rewardstypes.MsgSetContractMetadata
		expectError bool
		errorType   error
	}{
		{
			testCase: "err: empty request",
			prepare: func() *rewardstypes.MsgSetContractMetadata {
				return nil
			},
			expectError: true,
			errorType:   status.Error(codes.InvalidArgument, "empty request"),
		},
		{
			testCase: "err: invalid sender address",
			prepare: func() *rewardstypes.MsgSetContractMetadata {
				return &rewardstypes.MsgSetContractMetadata{
					SenderAddress: "👻",
					Metadata:      rewardstypes.ContractMetadata{},
				}
			},
			expectError: true,
			errorType:   fmt.Errorf("decoding bech32 failed: invalid bech32 string length 4"),
		},
		{
			testCase: "err: invalid contract address",
			prepare: func() *rewardstypes.MsgSetContractMetadata {
				accs, _ := e2eTesting.GenAccounts(1)
				return &rewardstypes.MsgSetContractMetadata{
					SenderAddress: accs[0].String(),
					Metadata: rewardstypes.ContractMetadata{
						ContractAddress: "👻",
					},
				}
			},
			expectError: true,
			errorType:   fmt.Errorf("decoding bech32 failed: invalid bech32 string length 4"),
		},
		{
			testCase: "err: contract does not exist",
			prepare: func() *rewardstypes.MsgSetContractMetadata {
				contractAddr := e2eTesting.GenContractAddresses(1)[0]
				accs, _ := e2eTesting.GenAccounts(1)
				return &rewardstypes.MsgSetContractMetadata{
					SenderAddress: accs[0].String(),
					Metadata: rewardstypes.ContractMetadata{
						ContractAddress: contractAddr.String(),
					},
				}
			},
			expectError: true,
			errorType:   rewardstypes.ErrContractNotFound,
		},
		{
			testCase: "ok: all good'",
			prepare: func() *rewardstypes.MsgSetContractMetadata {
				contractAdminAcc, otherAcc := s.chain.GetAccount(0), s.chain.GetAccount(1)
				contractViewer := testutils.NewMockContractViewer()
				s.chain.GetApp().RewardsKeeper.SetContractInfoViewer(contractViewer)
				contractAddr := e2eTesting.GenContractAddresses(1)[0]
				contractViewer.AddContractAdmin(contractAddr.String(), contractAdminAcc.Address.String())

				return &rewardstypes.MsgSetContractMetadata{
					SenderAddress: contractAdminAcc.Address.String(),
					Metadata: rewardstypes.ContractMetadata{
						ContractAddress: contractAddr.String(),
						OwnerAddress:    contractAdminAcc.Address.String(),
						RewardsAddress:  otherAcc.Address.String(),
					},
				}
			},
			expectError: false,
			errorType:   nil,
		},
	}

	server := keeper.NewMsgServer(s.chain.GetApp().RewardsKeeper)
	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case: %s", tc.testCase), func() {
			req := tc.prepare()
			res, err := server.SetContractMetadata(sdk.WrapSDKContext(s.chain.GetContext()), req)
			if tc.expectError {
				s.Require().Error(err)
				s.Require().Equal(tc.errorType.Error(), err.Error())
			} else {
				s.Require().NoError(err)
				s.Require().Equal(res, rewardstypes.MsgSetContractMetadataResponse{})
			}
		})
	}
}
