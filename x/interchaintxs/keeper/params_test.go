package keeper_test

import (
	"github.com/archway-network/archway/x/interchaintxs/types"
)

func (s *KeeperTestSuite) TestGetParams() {
	ctx, k := s.chain.GetContext().WithBlockHeight(100), s.chain.GetApp().Keepers.InterchainTxsKeeper
	params := types.DefaultParams()

	err := k.SetParams(ctx, params)
	s.Require().NoError(err)

	s.Require().EqualValues(params, k.GetParams(ctx))
}
