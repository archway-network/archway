package keeper_test

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	cwerrorsKeeper "github.com/archway-network/archway/x/cwerrors/keeper"
	"github.com/archway-network/archway/x/cwerrors/types"
)

func (s *KeeperTestSuite) TestSubscribeToError() {
	// Setting up chain and contract in mock wasm keeper
	ctx, keeper := s.chain.GetContext(), s.chain.GetApp().Keepers.CWErrorsKeeper
	contractViewer := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	contractAddr := e2eTesting.GenContractAddresses(2)[0]
	contractAddr2 := e2eTesting.GenContractAddresses(2)[1]
	contractAdminAcc := s.chain.GetAccount(2)
	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)
	params, err := keeper.GetParams(ctx)
	s.Require().NoError(err)
	params.SubscriptionFee = sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)
	err = keeper.SetParams(ctx, params)
	s.Require().NoError(err)

	expectedEndHeight := ctx.BlockHeight() + params.SubscriptionPeriod

	msgServer := cwerrorsKeeper.NewMsgServer(keeper)

	testCases := []struct {
		testCase    string
		input       func() *types.MsgSubscribeToError
		expectError bool
		errorType   error
	}{
		{
			testCase: "FAIL: empty request",
			input: func() *types.MsgSubscribeToError {
				return nil
			},
			expectError: true,
			errorType:   status.Error(codes.InvalidArgument, "empty request"),
		},
		{
			testCase: "FAIL: invalid sender address",
			input: func() *types.MsgSubscribeToError {
				return &types.MsgSubscribeToError{
					Sender:          "👻",
					ContractAddress: contractAddr.String(),
					Fee:             sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
				}
			},
			expectError: true,
			errorType:   errors.New("invalid bech32 string length 4"),
		},
		{
			testCase: "FAIL: invalid contract address",
			input: func() *types.MsgSubscribeToError {
				return &types.MsgSubscribeToError{
					Sender:          contractAdminAcc.Address.String(),
					ContractAddress: "👻",
					Fee:             sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
				}
			},
			expectError: true,
			errorType:   errors.New("invalid bech32 string length 4"),
		},
		{
			testCase: "FAIL: contract not found",
			input: func() *types.MsgSubscribeToError {
				return &types.MsgSubscribeToError{
					Sender:          contractAdminAcc.Address.String(),
					ContractAddress: contractAddr2.String(),
					Fee:             sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
				}
			},
			expectError: true,
			errorType:   types.ErrContractNotFound,
		},
		{
			testCase: "FAIL: account doesnt have enough balance",
			input: func() *types.MsgSubscribeToError {
				return &types.MsgSubscribeToError{
					Sender:          contractAddr.String(),
					ContractAddress: contractAddr.String(),
					Fee:             params.SubscriptionFee,
				}
			},
			expectError: true,
			errorType:   sdkerrors.ErrInsufficientFunds,
		},
		{
			testCase: "OK: valid request",
			input: func() *types.MsgSubscribeToError {
				return &types.MsgSubscribeToError{
					Sender:          contractAdminAcc.Address.String(),
					ContractAddress: contractAddr.String(),
					Fee:             params.SubscriptionFee,
				}
			},
			expectError: false,
		},
	}
	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case: %s", tc.testCase), func() {
			req := tc.input()
			res, err := msgServer.SubscribeToError(ctx, req)
			if tc.expectError {
				s.Require().Error(err)
				s.Assert().ErrorContains(err, tc.errorType.Error())
			} else {
				s.Require().NoError(err)
				s.Require().Equal(expectedEndHeight, res.SubscriptionValidTill)
			}
		})
	}
}

func (s *KeeperTestSuite) TestUpdateParams() {
	// Setting up chain and contract in mock wasm keeper
	ctx, keeper := s.chain.GetContext().WithBlockHeight(101), s.chain.GetApp().Keepers.CWErrorsKeeper
	contractViewer := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(2)
	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)
	govAddr := s.chain.GetApp().Keepers.AccountKeeper.GetModuleAccount(ctx, "gov").GetAddress()

	msgServer := cwerrorsKeeper.NewMsgServer(keeper)

	testCases := []struct {
		testCase    string
		input       func() *types.MsgUpdateParams
		expectError bool
		errorType   error
	}{
		{
			testCase: "FAIL: empty request",
			input: func() *types.MsgUpdateParams {
				return nil
			},
			expectError: true,
			errorType:   status.Error(codes.InvalidArgument, "empty request"),
		},
		{
			testCase: "FAIL: invalid authority address",
			input: func() *types.MsgUpdateParams {
				return &types.MsgUpdateParams{
					Authority: "👻",
					Params:    types.DefaultParams(),
				}
			},
			expectError: true,
			errorType:   errors.New("invalid bech32 string length 4"),
		},
		{
			testCase: "FAIL: unauthorized address",
			input: func() *types.MsgUpdateParams {
				return &types.MsgUpdateParams{
					Authority: contractAdminAcc.Address.String(),
					Params:    types.DefaultParams(),
				}
			},
			expectError: true,
			errorType:   types.ErrUnauthorized,
		},
		{
			testCase: "FAIL: invalid params",
			input: func() *types.MsgUpdateParams {
				return &types.MsgUpdateParams{
					Authority: govAddr.String(),
					Params: types.NewParams(
						0,
						sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
						100,
					),
				}
			},
			expectError: true,
			errorType:   errors.New("ErrorStoredTime must be greater than 0"),
		},
		{
			testCase: "OK: valid request",
			input: func() *types.MsgUpdateParams {
				return &types.MsgUpdateParams{
					Authority: govAddr.String(),
					Params:    types.DefaultParams(),
				}
			},
			expectError: false,
		},
	}
	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case: %s", tc.testCase), func() {
			req := tc.input()
			res, err := msgServer.UpdateParams(ctx, req)
			if tc.expectError {
				s.Require().Error(err)
				s.Assert().ErrorContains(err, tc.errorType.Error())
			} else {
				s.Require().NoError(err)
				s.Require().Equal(&types.MsgUpdateParamsResponse{}, res)
			}
		})
	}
}
