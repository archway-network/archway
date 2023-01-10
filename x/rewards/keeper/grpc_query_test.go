package keeper_test

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/rewards/keeper"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

func (s *KeeperTestSuite) TestGRPC_Params() {
	ctx, k := s.chain.GetContext(), s.chain.GetApp().RewardsKeeper
	querySrvr := keeper.NewQueryServer(k)
	params := rewardsTypes.Params{
		InflationRewardsRatio: sdk.MustNewDecFromStr("0.1"),
		TxFeeRebateRatio:      sdk.MustNewDecFromStr("0.1"),
		MaxWithdrawRecords:    uint64(2),
	}
	k.SetParams(ctx, params)

	s.Run("err: empty request", func() {
		_, err := querySrvr.Params(sdk.WrapSDKContext(ctx), nil)
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "empty request"), err)
	})

	s.Run("ok: gets params", func() {
		res, err := querySrvr.Params(sdk.WrapSDKContext(ctx), &rewardsTypes.QueryParamsRequest{})
		s.Require().NoError(err)
		s.Require().Equal(params.InflationRewardsRatio, res.Params.InflationRewardsRatio)
		s.Require().Equal(params.TxFeeRebateRatio, res.Params.TxFeeRebateRatio)
		s.Require().Equal(params.MaxWithdrawRecords, res.Params.MaxWithdrawRecords)
	})
}
