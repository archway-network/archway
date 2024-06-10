package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	callbackKeeper "github.com/archway-network/archway/x/callback/keeper"
	"github.com/archway-network/archway/x/callback/types"
)

func TestRequestCallback(t *testing.T) {
	// Setting up chain and contract in mock wasm keeper
	keeper, ctx := testutils.CallbackKeeper(t)
	wasmKeeper := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(wasmKeeper)

	contractAddresses := e2eTesting.GenContractAddresses(3)
	contractAddr := contractAddresses[0]
	contractAdminAcc := testutils.AccAddress()
	wasmKeeper.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.String(),
	)

	msgServer := callbackKeeper.NewMsgServer(keeper)

	testCases := []struct {
		testCase    string
		input       func() *types.MsgRequestCallback
		expectError bool
		errorType   error
	}{
		{
			testCase: "FAIL: empty request",
			input: func() *types.MsgRequestCallback {
				return nil
			},
			expectError: true,
			errorType:   status.Error(codes.InvalidArgument, "empty request"),
		},
		{
			testCase: "FAIL: insufficient callback fees",
			input: func() *types.MsgRequestCallback {
				return &types.MsgRequestCallback{
					ContractAddress: contractAddr.String(),
					JobId:           1,
					CallbackHeight:  102,
					Sender:          contractAddr.String(),
					Fees:            sdk.NewInt64Coin("stake", 0),
				}
			},
			expectError: true,
			errorType:   types.ErrInsufficientFees,
		},
		{
			testCase: "FAIL: contract does not exist",
			input: func() *types.MsgRequestCallback {
				return &types.MsgRequestCallback{
					ContractAddress: contractAdminAcc.String(),
					JobId:           1,
					CallbackHeight:  102,
					Sender:          contractAddr.String(),
					Fees:            sdk.NewInt64Coin("stake", 3500000000),
				}
			},
			expectError: true,
			errorType:   types.ErrContractNotFound,
		},
		{
			testCase: "Fail: account does not have enough balance",
			input: func() *types.MsgRequestCallback {
				return &types.MsgRequestCallback{
					ContractAddress: contractAddr.String(),
					JobId:           1,
					CallbackHeight:  102,
					Sender:          contractAddr.String(),
					Fees:            sdk.NewInt64Coin("stake", 3500000001),
				}
			},
			expectError: true,
			errorType:   sdkerrors.ErrInsufficientFunds,
		},
		{
			testCase: "OK: successfully register callback",
			input: func() *types.MsgRequestCallback {
				return &types.MsgRequestCallback{
					ContractAddress: contractAddr.String(),
					JobId:           1,
					CallbackHeight:  120,
					Sender:          contractAdminAcc.String(),
					Fees:            sdk.NewInt64Coin("stake", 3500000000),
				}
			},
			expectError: false,
			errorType:   nil,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case: %s", tc.testCase), func(t *testing.T) {
			req := tc.input()
			res, err := msgServer.RequestCallback(ctx, req)
			if tc.expectError {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.errorType)
			} else {
				require.NoError(t, err)
				require.Equal(t, &types.MsgRequestCallbackResponse{}, res)
				// Ensuring the callback exists now
				exists, err := keeper.ExistsCallback(ctx, req.CallbackHeight, req.ContractAddress, req.JobId)
				require.NoError(t, err)
				require.True(t, exists)
			}
		})
	}
}

func TestCancelCallback(t *testing.T) {
	// Setting up chain and contract in mock wasm keeper
	keeper, ctx := testutils.CallbackKeeper(t)
	wasmKeeper := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(wasmKeeper)

	contractAddresses := e2eTesting.GenContractAddresses(3)
	contractAddr := contractAddresses[0]
	contractAdminAcc := testutils.AccAddress()
	wasmKeeper.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.String(),
	)

	msgServer := callbackKeeper.NewMsgServer(keeper)
	// Setting up an existing callback to delete
	reqMsg := &types.MsgRequestCallback{
		ContractAddress: contractAddr.String(),
		JobId:           1,
		CallbackHeight:  130,
		Sender:          contractAdminAcc.String(),
		Fees:            sdk.NewInt64Coin("stake", 3500000000),
	}
	_, err := msgServer.RequestCallback(ctx, reqMsg)
	require.NoError(t, err)
	callback, err := keeper.GetCallback(ctx, reqMsg.CallbackHeight, reqMsg.ContractAddress, reqMsg.JobId)
	require.NoError(t, err)

	testCases := []struct {
		testCase    string
		input       func() *types.MsgCancelCallback
		expectError bool
		errorType   error
	}{
		{
			testCase: "FAIL: empty request",
			input: func() *types.MsgCancelCallback {
				return nil
			},
			expectError: true,
			errorType:   status.Error(codes.InvalidArgument, "empty request"),
		},
		{
			testCase: "FAIL: callback does not exist",
			input: func() *types.MsgCancelCallback {
				return &types.MsgCancelCallback{
					ContractAddress: contractAddr.String(),
					JobId:           2,
					CallbackHeight:  130,
					Sender:          contractAdminAcc.String(),
				}
			},
			expectError: true,
			errorType:   types.ErrCallbackNotFound,
		},
		{
			testCase: "FAIL: sender is not authorized to cancel callback",
			input: func() *types.MsgCancelCallback {
				return &types.MsgCancelCallback{
					ContractAddress: contractAddr.String(),
					JobId:           1,
					CallbackHeight:  130,
					Sender:          testutils.AccAddress().String(),
				}
			},
			expectError: true,
			errorType:   types.ErrUnauthorized,
		},
		{
			testCase: "OK: successfully cancel callback",
			input: func() *types.MsgCancelCallback {
				return &types.MsgCancelCallback{
					ContractAddress: contractAddr.String(),
					JobId:           1,
					CallbackHeight:  130,
					Sender:          contractAdminAcc.String(),
				}
			},
			expectError: false,
			errorType:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case: %s", tc.testCase), func(t *testing.T) {
			req := tc.input()
			res, err := msgServer.CancelCallback(ctx, req)
			if tc.expectError {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.errorType)
			} else {
				require.NoError(t, err)
				// Ensuring the callback no longer exists
				exists, err := keeper.ExistsCallback(ctx, req.CallbackHeight, req.ContractAddress, req.JobId)
				require.NoError(t, err)
				require.False(t, exists)
				// Ensuring the refund amount matches expected amount
				refundAmount := callback.FeeSplit.TransactionFees.Add(*callback.FeeSplit.SurplusFees)
				require.Equal(t, refundAmount, res.Refund)
			}
		})
	}
}
