package keeper_test

import (
	"fmt"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	callbackKeeper "github.com/archway-network/archway/x/callback/keeper"
	"github.com/archway-network/archway/x/callback/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *KeeperTestSuite) TestRequestCallback() {
	ctx, keeper := s.chain.GetContext().WithBlockHeight(100), s.chain.GetApp().Keepers.CallbackKeeper
	contractViewer := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	msgServer := callbackKeeper.NewMsgServer(keeper)

	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)

	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)
	s.chain.GetApp().Keepers.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("stake", 3500000000)))
	s.chain.GetApp().Keepers.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, contractAdminAcc.Address, sdk.NewCoins(sdk.NewInt64Coin("stake", 3500000000)))

	testCases := []struct {
		testCase    string
		prepare     func() *types.MsgRequestCallback
		expectError bool
		errorType   error
	}{
		{
			testCase: "FAIL: empty request",
			prepare: func() *types.MsgRequestCallback {
				return nil
			},
			expectError: true,
			errorType:   status.Error(codes.InvalidArgument, "empty request"),
		},
		{
			testCase: "FAIL: fees insufficient",
			prepare: func() *types.MsgRequestCallback {
				return &types.MsgRequestCallback{
					ContractAddress: contractAddr.String(),
					JobId:           1,
					CallbackHeight:  101,
					Sender:          contractAddr.String(),
					Fees:            sdk.NewInt64Coin("stake", 0),
				}
			},
			expectError: true,
			errorType:   types.ErrInsufficientFees,
		},
		{
			testCase: "FAIL: error saving as contract does not exist",
			prepare: func() *types.MsgRequestCallback {
				return &types.MsgRequestCallback{
					ContractAddress: contractAdminAcc.Address.String(),
					JobId:           1,
					CallbackHeight:  101,
					Sender:          contractAddr.String(),
					Fees:            sdk.NewInt64Coin("stake", 3500000000),
				}
			},
			expectError: true,
			errorType:   types.ErrContractNotFound,
		},
		{
			testCase: "Fail: account does not have enough balance",
			prepare: func() *types.MsgRequestCallback {
				return &types.MsgRequestCallback{
					ContractAddress: contractAddr.String(),
					JobId:           1,
					CallbackHeight:  101,
					Sender:          contractAddr.String(),
					Fees:            sdk.NewInt64Coin("stake", 3500000000),
				}
			},
			expectError: true,
			errorType:   sdkerrors.ErrInsufficientFunds,
		},
		{
			testCase: "OK: register callback",
			prepare: func() *types.MsgRequestCallback {

				return &types.MsgRequestCallback{
					ContractAddress: contractAddr.String(),
					JobId:           1,
					CallbackHeight:  111,
					Sender:          contractAdminAcc.Address.String(),
					Fees:            sdk.NewInt64Coin("stake", 3500000000),
				}
			},
			expectError: false,
			errorType:   nil,
		},
	}
	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case: %s", tc.testCase), func() {
			req := tc.prepare()
			res, err := msgServer.RequestCallback(sdk.WrapSDKContext(ctx), req)
			if tc.expectError {
				s.Require().Error(err)
				s.Assert().ErrorIs(err, tc.errorType)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(&types.MsgRequestCallbackResponse{}, res)

				callback, err := keeper.GetCallback(ctx, req.CallbackHeight, req.ContractAddress, req.JobId)
				s.Require().NoError(err)
				s.Require().NotEqual(types.Callback{}, callback)
			}
		})
	}
}

func (s *KeeperTestSuite) TestCancelCallback() {

	// request is nil
	// callback does not exist
	// successfully delete callback
}
