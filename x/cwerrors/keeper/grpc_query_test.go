package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	cwerrorsKeeper "github.com/archway-network/archway/x/cwerrors/keeper"
	"github.com/archway-network/archway/x/cwerrors/types"
)

func (s *KeeperTestSuite) TestErrors() {
	ctx, keeper := s.chain.GetContext(), s.chain.GetApp().Keepers.CWErrorsKeeper
	contractViewer := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	contractAddresses := e2eTesting.GenContractAddresses(3)
	contractAddr := contractAddresses[0]
	contractAddr2 := contractAddresses[1]
	contractAdminAcc := s.chain.GetAccount(0)
	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)
	contractViewer.AddContractAdmin(
		contractAddr2.String(),
		contractAdminAcc.Address.String(),
	)

	queryServer := cwerrorsKeeper.NewQueryServer(keeper)

	// Sending nil query
	_, err := queryServer.Errors(sdk.WrapSDKContext(ctx), nil)
	s.Require().Error(err)

	// Set errors for block 1
	// 2 errors for contract1
	// 1 error for contract2
	contract1Err := types.SudoError{
		ContractAddress: contractAddr.String(),
		ModuleName:      "test",
	}
	contract2Err := types.SudoError{
		ContractAddress: contractAddr2.String(),
		ModuleName:      "test",
	}
	err = keeper.SetError(ctx, contract1Err)
	s.Require().NoError(err)
	err = keeper.SetError(ctx, contract1Err)
	s.Require().NoError(err)
	err = keeper.SetError(ctx, contract2Err)
	s.Require().NoError(err)

	// Check number of errors match
	res, err := queryServer.Errors(sdk.WrapSDKContext(ctx), &types.QueryErrorsRequest{ContractAddress: contractAddr.String()})
	s.Require().NoError(err)
	s.Require().Len(res.Errors, 2)
	res, err = queryServer.Errors(sdk.WrapSDKContext(ctx), &types.QueryErrorsRequest{ContractAddress: contractAddr2.String()})
	s.Require().NoError(err)
	s.Require().Len(res.Errors, 1)

	// Increment block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// Set errors for block 2
	// 1 error for contract1
	// 1 error for contract2
	err = keeper.SetError(ctx, contract1Err)
	s.Require().NoError(err)
	err = keeper.SetError(ctx, contract2Err)
	s.Require().NoError(err)

	// Check number of errors match
	res, err = queryServer.Errors(sdk.WrapSDKContext(ctx), &types.QueryErrorsRequest{ContractAddress: contractAddr.String()})
	s.Require().NoError(err)
	s.Require().Len(res.Errors, 3)
	res, err = queryServer.Errors(sdk.WrapSDKContext(ctx), &types.QueryErrorsRequest{ContractAddress: contractAddr2.String()})
	s.Require().NoError(err)
	s.Require().Len(res.Errors, 2)
}

func (s *KeeperTestSuite) TestIsSubscribed() {
	ctx, keeper := s.chain.GetContext(), s.chain.GetApp().Keepers.CWErrorsKeeper
	contractViewer := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	contractAddr := e2eTesting.GenContractAddresses(2)[0]
	contractAdminAcc := s.chain.GetAccount(2)
	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)
	queryServer := cwerrorsKeeper.NewQueryServer(keeper)

	// TEST CASE 1: empty request
	_, err := queryServer.IsSubscribed(sdk.WrapSDKContext(ctx), nil)
	s.Require().Error(err)

	// TEST CASE 2: invalid contract address
	_, err = queryServer.IsSubscribed(sdk.WrapSDKContext(ctx), &types.QueryIsSubscribedRequest{ContractAddress: "👻"})
	s.Require().Error(err)

	// TEST CASE 3: subscription not found
	res, err := queryServer.IsSubscribed(sdk.WrapSDKContext(ctx), &types.QueryIsSubscribedRequest{ContractAddress: contractAddr.String()})
	s.Require().NoError(err)
	s.Require().False(res.Subscribed)

	// TEST CASE 4: subscription found
	expectedEndHeight, err := keeper.SetSubscription(ctx, contractAdminAcc.Address, contractAddr, sdk.NewInt64Coin(sdk.DefaultBondDenom, 0))
	s.Require().NoError(err)
	res, err = queryServer.IsSubscribed(sdk.WrapSDKContext(ctx), &types.QueryIsSubscribedRequest{ContractAddress: contractAddr.String()})
	s.Require().NoError(err)
	s.Require().True(res.Subscribed)
	s.Require().Equal(expectedEndHeight, res.SubscriptionValidTill)
}

func (s *KeeperTestSuite) TestParams() {
	ctx, keeper := s.chain.GetContext(), s.chain.GetApp().Keepers.CWErrorsKeeper
	queryServer := cwerrorsKeeper.NewQueryServer(keeper)

	// Sending nil query
	_, err := queryServer.Params(sdk.WrapSDKContext(ctx), nil)
	s.Require().Error(err)

	// Set params
	params := types.Params{
		ErrorStoredTime:    100,
		SubscriptionFee:    sdk.NewInt64Coin(sdk.DefaultBondDenom, 2),
		SubscriptionPeriod: 100,
	}
	keeper.SetParams(ctx, params)

	// Query params
	res, err := queryServer.Params(sdk.WrapSDKContext(ctx), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(params, res.Params)
}
