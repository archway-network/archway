package keeper_test

import (
	"fmt"
	"testing"

	math "cosmossdk.io/math"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/rewards/keeper"
	rewardstypes "github.com/archway-network/archway/x/rewards/types"
)

func TestMsgServer_SetContractMetadata(t *testing.T) {
	k, ctx, _ := testutils.RewardsKeeper(t)
	wk := testutils.NewMockContractViewer()
	k.SetContractInfoViewer(wk)
	contractAdminAcc, otherAcc := testutils.AccAddress(), testutils.AccAddress()
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
					SenderAddress: contractAdminAcc.String(),
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
					SenderAddress: contractAdminAcc.String(),
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
				wk.AddContractAdmin(contractAddr.String(), contractAdminAcc.String())

				return &rewardstypes.MsgSetContractMetadata{
					SenderAddress: otherAcc.String(),
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
				wk.AddContractAdmin(contractAddr.String(), contractAdminAcc.String())

				return &rewardstypes.MsgSetContractMetadata{
					SenderAddress: contractAdminAcc.String(),
					Metadata: rewardstypes.ContractMetadata{
						ContractAddress: contractAddr.String(),
						OwnerAddress:    contractAdminAcc.String(),
						RewardsAddress:  otherAcc.String(),
					},
				}
			},
			expectError: false,
			errorType:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case: %s", tc.testCase), func(t *testing.T) {
			req := tc.prepare()
			res, err := server.SetContractMetadata(ctx, req)
			if tc.expectError {
				require.Error(t, err)
				require.Equal(t, tc.errorType.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, &rewardstypes.MsgSetContractMetadataResponse{}, res)
			}
		})
	}
}

func TestMsgServer_WithdrawRewards(t *testing.T) {

	k, ctx, _ := testutils.RewardsKeeper(t)
	acc := testutils.AccAddress()

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
		t.Run(fmt.Sprintf("Case: %s", tc.testCase), func(t *testing.T) {
			req := tc.prepare()
			res, err := server.WithdrawRewards(ctx, req)
			if tc.expectError {
				require.Error(t, err)
				require.Equal(t, tc.errorType.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.EqualValues(t, 0, res.RecordsNum)
				require.Empty(t, res.TotalRewards)
			}
		})
	}
}

func TestMsgServer_SetFlatFee(t *testing.T) {
	k, ctx, _ := testutils.RewardsKeeper(t)
	wk := testutils.NewMockContractViewer()
	k.SetContractInfoViewer(wk)
	contractAdminAcc, otherAcc := testutils.AccAddress(), testutils.AccAddress()
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
					SenderAddress:   contractAdminAcc.String(),
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
					SenderAddress:   contractAdminAcc.String(),
					ContractAddress: contractAddr.String(),
				}
			},
			expectError: true,
			errorType:   rewardstypes.ErrMetadataNotFound,
		},
		{
			testCase: "err: the message sender is not the contract owner",
			prepare: func() *rewardstypes.MsgSetFlatFee {
				wk.AddContractAdmin(contractAddr.String(), contractAdminAcc.String())
				contractMetadata := rewardstypes.ContractMetadata{
					ContractAddress: contractAddr.String(),
					OwnerAddress:    contractAdminAcc.String(),
					RewardsAddress:  otherAcc.String(),
				}
				err := k.SetContractMetadata(ctx, contractAdminAcc, contractAddr, contractMetadata)
				require.NoError(t, err)

				return &rewardstypes.MsgSetFlatFee{
					SenderAddress:   otherAcc.String(),
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
				wk.AddContractAdmin(contractAddr.String(), contractAdminAcc.String())
				contractMetadata := rewardstypes.ContractMetadata{
					ContractAddress: contractAddr.String(),
					OwnerAddress:    contractAdminAcc.String(),
					RewardsAddress:  otherAcc.String(),
				}
				err := k.SetContractMetadata(ctx, contractAdminAcc, contractAddr, contractMetadata)
				require.NoError(t, err)

				return &rewardstypes.MsgSetFlatFee{
					SenderAddress:   contractAdminAcc.String(),
					ContractAddress: contractAddr.String(),
					FlatFeeAmount:   sdk.NewInt64Coin("token", 10),
				}
			},
			expectError: false,
			errorType:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case: %s", tc.testCase), func(t *testing.T) {
			req := tc.prepare()
			res, err := server.SetFlatFee(ctx, req)
			if tc.expectError {
				require.Error(t, err)
				require.Equal(t, tc.errorType.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, &rewardstypes.MsgSetFlatFeeResponse{}, res)
			}
		})
	}
}

func TestMsgServer_UpdateParams(t *testing.T) {
	k, ctx, _ := testutils.RewardsKeeper(t)
	account := testutils.AccAddress()

	server := keeper.NewMsgServer(k)

	govAddress := "cosmos1a48wdtjn3egw7swhfkeshwdtjvs6hq9nlyrwut"

	testCases := []struct {
		testCase    string
		prepare     func() *rewardstypes.MsgUpdateParams
		expectError bool
	}{
		{
			testCase: "fail: invalid params",
			prepare: func() *rewardstypes.MsgUpdateParams {
				params := rewardstypes.DefaultParams()
				params.InflationRewardsRatio = math.LegacyNewDecWithPrec(-2, 2)
				return &rewardstypes.MsgUpdateParams{
					Authority: govAddress,
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
					Authority: account.String(),
					Params:    rewardstypes.DefaultParams(),
				}
			},
			expectError: true,
		},
		{
			testCase: "ok: valid params with x/gov address",
			prepare: func() *rewardstypes.MsgUpdateParams {
				return &rewardstypes.MsgUpdateParams{
					Authority: govAddress,
					Params:    rewardstypes.DefaultParams(),
				}
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case: %s", tc.testCase), func(t *testing.T) {
			req := tc.prepare()
			res, err := server.UpdateParams(ctx, req)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, &rewardstypes.MsgUpdateParamsResponse{}, res)
			}
		})
	}
}
