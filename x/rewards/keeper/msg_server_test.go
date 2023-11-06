package keeper_test

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/rewards/keeper"
	rewardstypes "github.com/archway-network/archway/x/rewards/types"
)

func (s *KeeperTestSuite) TestMsgServer_SetContractMetadata() {
	ctx, k := s.chain.GetContext(), s.chain.GetApp().Keepers.RewardsKeeper
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
			errorType:   errorsmod.Wrap(rewardstypes.ErrUnauthorized, "metadata can only be created by the contract admin"),
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
				s.Require().Equal(&rewardstypes.MsgSetContractMetadataResponse{}, res)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgServer_WithdrawRewards() {
	ctx, k := s.chain.GetContext(), s.chain.GetApp().Keepers.RewardsKeeper
	acc := s.chain.GetAccount(0).Address

	server := keeper.NewMsgServer(k)

	testCases := []struct {
		testCase    string
		prepare     func() *rewardstypes.MsgWithdrawRewards
		expectError bool
		errorType   error
	}{
		{
			testCase: "err: empty request",
			prepare: func() *rewardstypes.MsgWithdrawRewards {
				return nil
			},
			expectError: true,
			errorType:   status.Error(codes.InvalidArgument, "empty request"),
		},
		{
			testCase: "err: invalid sender address",
			prepare: func() *rewardstypes.MsgWithdrawRewards {
				return &rewardstypes.MsgWithdrawRewards{
					RewardsAddress: "👻",
				}
			},
			expectError: true,
			errorType:   fmt.Errorf("decoding bech32 failed: invalid bech32 string length 4"),
		},
		{
			testCase: "err: rewards mode invalid",
			prepare: func() *rewardstypes.MsgWithdrawRewards {
				return &rewardstypes.MsgWithdrawRewards{
					RewardsAddress: acc.String(),
					Mode:           nil,
				}
			},
			expectError: true,
			errorType:   status.Error(codes.InvalidArgument, "invalid request mode"),
		},
		{
			testCase: "err: withdraw rewards records limit invalid",
			prepare: func() *rewardstypes.MsgWithdrawRewards {
				return &rewardstypes.MsgWithdrawRewards{
					RewardsAddress: acc.String(),
					Mode: &rewardstypes.MsgWithdrawRewards_RecordsLimit_{
						RecordsLimit: &rewardstypes.MsgWithdrawRewards_RecordsLimit{
							Limit: rewardstypes.MaxWithdrawRecordsParamLimit + 1,
						},
					},
				}
			},
			expectError: true,
			errorType:   status.Error(codes.InvalidArgument, errorsmod.Wrapf(rewardstypes.ErrInvalidRequest, "max withdraw records (25000) exceeded").Error()),
		},
		{
			testCase: "ok: withdraw rewards by records limit",
			prepare: func() *rewardstypes.MsgWithdrawRewards {
				return &rewardstypes.MsgWithdrawRewards{
					RewardsAddress: acc.String(),
					Mode: &rewardstypes.MsgWithdrawRewards_RecordsLimit_{
						RecordsLimit: &rewardstypes.MsgWithdrawRewards_RecordsLimit{
							Limit: 1,
						},
					},
				}
			},
			expectError: false,
			errorType:   nil,
		},
		{
			testCase: "ok: withdraw rewards by record ids",
			prepare: func() *rewardstypes.MsgWithdrawRewards {
				return &rewardstypes.MsgWithdrawRewards{
					RewardsAddress: acc.String(),
					Mode: &rewardstypes.MsgWithdrawRewards_RecordIds{
						RecordIds: &rewardstypes.MsgWithdrawRewards_RecordIDs{
							Ids: []uint64{},
						},
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
			res, err := server.WithdrawRewards(sdk.WrapSDKContext(ctx), req)
			if tc.expectError {
				s.Require().Error(err)
				s.Require().Equal(tc.errorType.Error(), err.Error())
			} else {
				s.Require().NoError(err)
				s.Require().EqualValues(0, res.RecordsNum)
				s.Require().Empty(res.TotalRewards)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgServer_SetFlatFee() {
	ctx, k := s.chain.GetContext(), s.chain.GetApp().Keepers.RewardsKeeper
	contractAdminAcc, otherAcc := s.chain.GetAccount(0), s.chain.GetAccount(1)
	contractViewer := testutils.NewMockContractViewer()
	k.SetContractInfoViewer(contractViewer)
	contractAddr := e2eTesting.GenContractAddresses(1)[0]

	server := keeper.NewMsgServer(k)

	testCases := []struct {
		testCase    string
		prepare     func() *rewardstypes.MsgSetFlatFee
		expectError bool
		errorType   error
	}{
		{
			testCase: "err: empty request",
			prepare: func() *rewardstypes.MsgSetFlatFee {
				return nil
			},
			expectError: true,
			errorType:   status.Error(codes.InvalidArgument, "empty request"),
		},
		{
			testCase: "err: invalid sender address",
			prepare: func() *rewardstypes.MsgSetFlatFee {
				return &rewardstypes.MsgSetFlatFee{
					SenderAddress: "👻",
				}
			},
			expectError: true,
			errorType:   fmt.Errorf("decoding bech32 failed: invalid bech32 string length 4"),
		},
		{
			testCase: "err: invalid contract address",
			prepare: func() *rewardstypes.MsgSetFlatFee {
				return &rewardstypes.MsgSetFlatFee{
					SenderAddress:   contractAdminAcc.Address.String(),
					ContractAddress: "👻",
				}
			},
			expectError: true,
			errorType:   fmt.Errorf("decoding bech32 failed: invalid bech32 string length 4"),
		},
		{
			testCase: "err: contract metadata not exist",
			prepare: func() *rewardstypes.MsgSetFlatFee {
				return &rewardstypes.MsgSetFlatFee{
					SenderAddress:   contractAdminAcc.Address.String(),
					ContractAddress: contractAddr.String(),
				}
			},
			expectError: true,
			errorType:   rewardstypes.ErrMetadataNotFound,
		},
		{
			testCase: "err: the message sender is not the contract owner",
			prepare: func() *rewardstypes.MsgSetFlatFee {
				contractViewer.AddContractAdmin(contractAddr.String(), contractAdminAcc.Address.String())
				contractMetadata := rewardstypes.ContractMetadata{
					ContractAddress: contractAddr.String(),
					OwnerAddress:    contractAdminAcc.Address.String(),
					RewardsAddress:  otherAcc.Address.String(),
				}
				err := k.SetContractMetadata(ctx, contractAdminAcc.Address, contractAddr, contractMetadata)
				s.Require().NoError(err)

				return &rewardstypes.MsgSetFlatFee{
					SenderAddress:   otherAcc.Address.String(),
					ContractAddress: contractAddr.String(),
					FlatFeeAmount:   sdk.NewInt64Coin("token", 10),
				}
			},
			expectError: true,
			errorType:   errorsmod.Wrap(rewardstypes.ErrUnauthorized, "flat_fee can only be set or changed by the contract owner"),
		},
		{
			testCase: "ok: all good'",
			prepare: func() *rewardstypes.MsgSetFlatFee {
				contractViewer.AddContractAdmin(contractAddr.String(), contractAdminAcc.Address.String())
				contractMetadata := rewardstypes.ContractMetadata{
					ContractAddress: contractAddr.String(),
					OwnerAddress:    contractAdminAcc.Address.String(),
					RewardsAddress:  otherAcc.Address.String(),
				}
				err := k.SetContractMetadata(ctx, contractAdminAcc.Address, contractAddr, contractMetadata)
				s.Require().NoError(err)

				return &rewardstypes.MsgSetFlatFee{
					SenderAddress:   contractAdminAcc.Address.String(),
					ContractAddress: contractAddr.String(),
					FlatFeeAmount:   sdk.NewInt64Coin("token", 10),
				}
			},
			expectError: false,
			errorType:   nil,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case: %s", tc.testCase), func() {
			req := tc.prepare()
			res, err := server.SetFlatFee(sdk.WrapSDKContext(ctx), req)
			if tc.expectError {
				s.Require().Error(err)
				s.Require().Equal(tc.errorType.Error(), err.Error())
			} else {
				s.Require().NoError(err)
				s.Require().Equal(&rewardstypes.MsgSetFlatFeeResponse{}, res)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgServer_UpdateParams() {
	ctx, k := s.chain.GetContext(), s.chain.GetApp().Keepers.RewardsKeeper
	account := s.chain.GetAccount(0)

	server := keeper.NewMsgServer(k)

	govAddress := s.chain.GetApp().Keepers.AccountKeeper.GetModuleAddress(govtypes.ModuleName)

	testCases := []struct {
		testCase    string
		prepare     func() *rewardstypes.MsgUpdateParams
		expectError bool
	}{
		{
			testCase: "fail: invalid params",
			prepare: func() *rewardstypes.MsgUpdateParams {
				params := rewardstypes.DefaultParams()
				params.InflationRewardsRatio = sdk.NewDecWithPrec(-2, 2)
				return &rewardstypes.MsgUpdateParams{
					Authority: govAddress.String(),
					Params:    params,
				}
			},
			expectError: true,
		},
		{
			testCase: "fail: invalid authority address",
			prepare: func() *rewardstypes.MsgUpdateParams {
				return &rewardstypes.MsgUpdateParams{
					Authority: "👻",
					Params:    rewardstypes.DefaultParams(),
				}
			},
			expectError: true,
		},
		{
			testCase: "fail: authority address is not gov address",
			prepare: func() *rewardstypes.MsgUpdateParams {
				return &rewardstypes.MsgUpdateParams{
					Authority: account.Address.String(),
					Params:    rewardstypes.DefaultParams(),
				}
			},
			expectError: true,
		},
		{
			testCase: "ok: valid params with x/gov address",
			prepare: func() *rewardstypes.MsgUpdateParams {
				return &rewardstypes.MsgUpdateParams{
					Authority: govAddress.String(),
					Params:    rewardstypes.DefaultParams(),
				}
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case: %s", tc.testCase), func() {
			req := tc.prepare()
			res, err := server.UpdateParams(sdk.WrapSDKContext(ctx), req)
			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(&rewardstypes.MsgUpdateParamsResponse{}, res)
			}
		})
	}
}
