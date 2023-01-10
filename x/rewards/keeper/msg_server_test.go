package keeper_test

import (
	"fmt"
	"github.com/archway-network/archway/pkg/testutils"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/rewards/keeper"
	rewardstypes "github.com/archway-network/archway/x/rewards/types"
)

func (s *KeeperTestSuite) TestMsgServer_SetContractMetadata() {
	ctx, k := s.chain.GetContext(), s.chain.GetApp().RewardsKeeper
	contractAdminAcc, otherAcc := s.chain.GetAccount(0), s.chain.GetAccount(1)
	contractViewer := testutils.NewMockContractViewer()
	k.SetContractInfoViewer(contractViewer)
	contractAddr := e2eTesting.GenContractAddresses(1)[0]

	server := keeper.NewMsgServer(k)

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
				return &rewardstypes.MsgSetContractMetadata{
					SenderAddress: contractAdminAcc.Address.String(),
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
				return &rewardstypes.MsgSetContractMetadata{
					SenderAddress: contractAdminAcc.Address.String(),
					Metadata: rewardstypes.ContractMetadata{
						ContractAddress: contractAddr.String(),
					},
				}
			},
			expectError: true,
			errorType:   rewardstypes.ErrContractNotFound,
		},
		{
			testCase: "err: the message sender is not the contract admin",
			prepare: func() *rewardstypes.MsgSetContractMetadata {
				contractViewer.AddContractAdmin(contractAddr.String(), contractAdminAcc.Address.String())

				return &rewardstypes.MsgSetContractMetadata{
					SenderAddress: otherAcc.Address.String(),
					Metadata: rewardstypes.ContractMetadata{
						ContractAddress: contractAddr.String(),
					},
				}
			},
			expectError: true,
			errorType:   sdkErrors.Wrap(rewardstypes.ErrUnauthorized, "metadata can only be created by the contract admin"),
		},
		{
			testCase: "ok: all good'",
			prepare: func() *rewardstypes.MsgSetContractMetadata {
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

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case: %s", tc.testCase), func() {
			req := tc.prepare()
			res, err := server.SetContractMetadata(sdk.WrapSDKContext(ctx), req)
			if tc.expectError {
				s.Require().Error(err)
				s.Require().Equal(tc.errorType.Error(), err.Error())
			} else {
				s.Require().NoError(err)
				s.Require().Equal(res, &rewardstypes.MsgSetContractMetadataResponse{})
			}
		})
	}
}
