package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	callbackKeeper "github.com/archway-network/archway/x/callback/keeper"
	"github.com/archway-network/archway/x/callback/types"
)

func (s *KeeperTestSuite) TestCallbacks() {
	ctx, keeper := s.chain.GetContext().WithBlockHeight(101), s.chain.GetApp().Keepers.CallbackKeeper
	contractViewer := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	queryServer := callbackKeeper.NewQueryServer(keeper)

	validCoin := sdk.NewInt64Coin("stake", 10)

	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)

	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)
	callbackHeight := int64(102)

	callback := types.Callback{
		ContractAddress: contractAddr.String(),
		JobId:           1,
		CallbackHeight:  callbackHeight,
		ReservedBy:      contractAddr.String(),
		FeeSplit: &types.CallbackFeesFeeSplit{
			TransactionFees:       &validCoin,
			BlockReservationFees:  &validCoin,
			FutureReservationFees: &validCoin,
			SurplusFees:           &validCoin,
		},
	}
	err := keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)

	callback.JobId = 2
	err = keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)

	callback.JobId = 3
	err = keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)

	callback.JobId = 4
	callback.CallbackHeight = callbackHeight + 1
	err = keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)

	testCases := []struct {
		testCase    string
		prepare     func() *types.QueryCallbacksRequest
		expectError bool
		response    int
	}{
		{
			testCase: "FAIL: empty request",
			prepare: func() *types.QueryCallbacksRequest {
				return nil
			},
			expectError: true,
			response:    0,
		},
		{
			testCase: "OK: no callbacks at requested height",
			prepare: func() *types.QueryCallbacksRequest {
				return &types.QueryCallbacksRequest{
					BlockHeight: 100,
				}
			},
			expectError: false,
			response:    0,
		},
		{
			testCase: "OK: get callbacks at requested height",
			prepare: func() *types.QueryCallbacksRequest {
				return &types.QueryCallbacksRequest{
					BlockHeight: callbackHeight,
				}
			},
			expectError: false,
			response:    3,
		},
		{
			testCase: "OK: get callbacks at requested height",
			prepare: func() *types.QueryCallbacksRequest {
				return &types.QueryCallbacksRequest{
					BlockHeight: callbackHeight + 1,
				}
			},
			expectError: false,
			response:    1,
		},
	}
	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case: %s", tc.testCase), func() {
			req := tc.prepare()
			res, err := queryServer.Callbacks(sdk.WrapSDKContext(ctx), req)
			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.response, len(res.Callbacks))
			}
		})
	}
}
