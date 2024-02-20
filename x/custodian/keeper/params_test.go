package keeper_test

import (
	"github.com/archway-network/archway/x/custodian/types"
)

func (s *KeeperTestSuite) TestGetParams() {
	ctx, k := s.chain.GetContext().WithBlockHeight(100), s.chain.GetApp().Keepers.CustodianKeeper
	params := types.DefaultParams()

	err := k.SetParams(ctx, params)
	s.Require().NoError(err)

	s.Require().EqualValues(params, k.GetParams(ctx))
}
