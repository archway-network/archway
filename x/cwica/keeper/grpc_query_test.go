package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/cwica/types"
)

// TestKeeper_Params tests the Params gRPC query method
func (s *KeeperTestSuite) TestParamsQuery() {
	ctx, keeper := s.chain.GetContext().WithBlockHeight(1), s.chain.GetApp().Keepers.CWICAKeeper
	wctx := sdk.WrapSDKContext(ctx)
	params := types.DefaultParams()
	err := keeper.SetParams(ctx, params)
	s.Require().NoError(err)

	// TEST CASE 1: invalid request
	response, err := keeper.Params(wctx, nil)
	s.Require().Error(err)
	s.Require().Nil(response)

	// TEST CASE 2: successfully fetched the params
	response, err = keeper.Params(wctx, &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(&types.QueryParamsResponse{Params: params}, response)
}
