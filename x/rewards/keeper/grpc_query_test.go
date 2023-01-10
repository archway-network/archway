package keeper_test

import (
	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
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

func (s *KeeperTestSuite) TestGRPC_ContractMetadata() {
	ctx, k := s.chain.GetContext(), s.chain.GetApp().RewardsKeeper
	querySrvr := keeper.NewQueryServer(k)
	contractViewer := testutils.NewMockContractViewer()
	k.SetContractInfoViewer(contractViewer)
	contractAddr := e2eTesting.GenContractAddresses(2)
	contractAdminAcc := s.chain.GetAccount(0)
	contractViewer.AddContractAdmin(contractAddr[0].String(), contractAdminAcc.Address.String())
	contractMeta := rewardsTypes.ContractMetadata{
		ContractAddress: contractAddr[0].String(),
		OwnerAddress:    contractAdminAcc.Address.String(),
	}
	err := k.SetContractMetadata(ctx, contractAdminAcc.Address, contractAddr[0], contractMeta)
	s.Require().NoError(err)

	s.Run("err: empty request", func() {
		_, err := querySrvr.ContractMetadata(sdk.WrapSDKContext(ctx), nil)
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "empty request"), err)
	})

	s.Run("err: invalid contract address", func() {
		_, err := querySrvr.ContractMetadata(sdk.WrapSDKContext(ctx), &rewardsTypes.QueryContractMetadataRequest{ContractAddress: "👻"})
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "invalid contract address: decoding bech32 failed: invalid bech32 string length 4"), err)
	})

	s.Run("err: contract metadata not found", func() {
		_, err := querySrvr.ContractMetadata(sdk.WrapSDKContext(ctx), &rewardsTypes.QueryContractMetadataRequest{ContractAddress: contractAddr[1].String()})
		s.Require().Error(err)
		s.Require().Equal(status.Errorf(codes.NotFound, "metadata for the contract: not found"), err)
	})

	s.Run("ok: gets contract metadata", func() {
		res, err := querySrvr.ContractMetadata(sdk.WrapSDKContext(ctx), &rewardsTypes.QueryContractMetadataRequest{ContractAddress: contractAddr[0].String()})
		s.Require().NoError(err)
		s.Require().Equal(contractMeta.ContractAddress, res.Metadata.ContractAddress)
		s.Require().Equal(contractMeta.RewardsAddress, res.Metadata.RewardsAddress)
		s.Require().Equal(contractMeta.OwnerAddress, res.Metadata.OwnerAddress)
	})
}

func (s *KeeperTestSuite) TestGRPC_BlockRewardsTracking() {
	ctx, k := s.chain.GetContext(), s.chain.GetApp().RewardsKeeper
	querySrvr := keeper.NewQueryServer(k)

	s.Run("err: empty request", func() {
		_, err := querySrvr.BlockRewardsTracking(sdk.WrapSDKContext(ctx), nil)
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "empty request"), err)
	})

	s.Run("ok: gets block rewards tracking", func() {
		res, err := querySrvr.BlockRewardsTracking(sdk.WrapSDKContext(ctx), &rewardsTypes.QueryBlockRewardsTrackingRequest{})
		s.Require().NoError(err)
		s.Require().Equal(0, len(res.Block.TxRewards))
		s.Require().Equal(ctx.BlockHeight(), res.Block.InflationRewards.Height)
	})
}
