package keeper_test

import (
	"github.com/archway-network/archway/x/cwica/types"
)

// TestGetParams tests the GetParams method of the CWICAKeeper
func (s *KeeperTestSuite) TestGetParams() {
	ctx, k := s.chain.GetContext().WithBlockHeight(100), s.chain.GetApp().Keepers.CWICAKeeper
	params := types.DefaultParams()

	err := k.SetParams(ctx, params)
	s.Require().NoError(err)

	p, err := k.GetParams(ctx)
	s.Require().NoError(err)
	s.Require().EqualValues(params, p)
}
