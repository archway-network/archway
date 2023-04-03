package keeper_test

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/rewards/keeper"
	"github.com/archway-network/archway/x/rewards/types"
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

func (s *KeeperTestSuite) TestGRPC_RewardsPool() {
	ctx, k := s.chain.GetContext(), s.chain.GetApp().RewardsKeeper
	querySrvr := keeper.NewQueryServer(k)

	s.Run("err: empty request", func() {
		_, err := querySrvr.RewardsPool(sdk.WrapSDKContext(ctx), nil)
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "empty request"), err)
	})

	s.Run("ok: gets rewards pool", func() {
		res, err := querySrvr.RewardsPool(sdk.WrapSDKContext(ctx), &rewardsTypes.QueryRewardsPoolRequest{})
		s.Require().NoError(err)
		s.Require().NotNil(res)
	})
}

func (s *KeeperTestSuite) TestGRPC_EstimateTxFees() {
	ctx, k := s.chain.GetContext(), s.chain.GetApp().RewardsKeeper

	querySrvr := keeper.NewQueryServer(k)

	s.Run("err: empty request", func() {
		_, err := querySrvr.EstimateTxFees(sdk.WrapSDKContext(ctx), nil)
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "empty request"), err)
	})

	s.Run("ok: gets estimated tx fees", func() {
		expectedFee := sdk.NewInt64Coin("stake", 0)
		res, err := querySrvr.EstimateTxFees(sdk.WrapSDKContext(ctx), &rewardsTypes.QueryEstimateTxFeesRequest{GasLimit: 0})
		s.Require().NoError(err)
		s.Require().NotNil(res)
		fees := sdk.NewCoins(res.EstimatedFee...)
		s.Require().EqualValues(expectedFee.Amount, fees.AmountOf("stake"))
	})

	minConsFee := sdk.NewInt64Coin("stake", 100)
	s.Run("ok: gets estimated tx fees (custom minconsfee set)", func() {
		s.chain.GetApp().RewardsKeeper.GetState().MinConsensusFee(ctx).SetFee(sdk.NewDecCoinFromCoin(minConsFee))
		res, err := querySrvr.EstimateTxFees(sdk.WrapSDKContext(ctx), &rewardsTypes.QueryEstimateTxFeesRequest{GasLimit: 1})
		s.Require().NoError(err)
		s.Require().NotNil(res)
		fees := sdk.NewCoins(res.EstimatedFee...)
		s.Require().EqualValues(minConsFee.Amount, fees.AmountOf("stake"))
	})

	s.Run("ok: gets estimated tx fees inclulding contract flat fee(diff denom)", func() {
		expectedFlatFee := sdk.NewInt64Coin("token", 123)
		contractAdminAcc := s.chain.GetAccount(0)
		contractViewer := testutils.NewMockContractViewer()
		k.SetContractInfoViewer(contractViewer)
		contractAddr := e2eTesting.GenContractAddresses(1)[0]
		contractViewer.AddContractAdmin(contractAddr.String(), contractAdminAcc.Address.String())
		err := k.SetContractMetadata(ctx, contractAdminAcc.Address, contractAddr, rewardsTypes.ContractMetadata{
			ContractAddress: contractAddr.String(),
			OwnerAddress:    contractAdminAcc.Address.String(),
		})
		s.Require().NoError(err)
		err = k.SetFlatFee(ctx, contractAdminAcc.Address, types.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         expectedFlatFee,
		})
		s.Require().NoError(err)

		res, err := querySrvr.EstimateTxFees(sdk.WrapSDKContext(ctx), &rewardsTypes.QueryEstimateTxFeesRequest{GasLimit: 1, ContractAddress: contractAddr.String()})
		s.Require().NoError(err)
		s.Require().NotNil(res)
		fees := sdk.NewCoins(res.EstimatedFee...)
		s.Require().Equal(expectedFlatFee.Amount, fees.AmountOf("token"))
		s.Require().EqualValues(minConsFee.Amount, fees.AmountOf("stake"))
	})

	s.Run("ok: gets estimated tx fees inclulding contract flat fee(same denom)", func() {
		expectedFlatFee := sdk.NewInt64Coin("stake", 123)
		contractAdminAcc := s.chain.GetAccount(0)
		contractViewer := testutils.NewMockContractViewer()
		k.SetContractInfoViewer(contractViewer)
		contractAddr := e2eTesting.GenContractAddresses(1)[0]
		contractViewer.AddContractAdmin(contractAddr.String(), contractAdminAcc.Address.String())
		err := k.SetContractMetadata(ctx, contractAdminAcc.Address, contractAddr, rewardsTypes.ContractMetadata{
			ContractAddress: contractAddr.String(),
			OwnerAddress:    contractAdminAcc.Address.String(),
		})
		s.Require().NoError(err)
		err = k.SetFlatFee(ctx, contractAdminAcc.Address, types.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         expectedFlatFee,
		})
		s.Require().NoError(err)

		res, err := querySrvr.EstimateTxFees(sdk.WrapSDKContext(ctx), &rewardsTypes.QueryEstimateTxFeesRequest{GasLimit: 1, ContractAddress: contractAddr.String()})
		s.Require().NoError(err)
		s.Require().NotNil(res)
		fees := sdk.NewCoins(res.EstimatedFee...)
		s.Require().Equal(expectedFlatFee.Add(minConsFee).Amount, fees.AmountOf("stake"))
	})
}

func (s *KeeperTestSuite) TestGRPC_OutstandingRewards() {
	ctx, k := s.chain.GetContext(), s.chain.GetApp().RewardsKeeper

	querySrvr := keeper.NewQueryServer(k)

	s.Run("err: empty request", func() {
		_, err := querySrvr.OutstandingRewards(sdk.WrapSDKContext(ctx), nil)
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "empty request"), err)
	})

	s.Run("err: invalid rewards address", func() {
		_, err := querySrvr.OutstandingRewards(sdk.WrapSDKContext(ctx), &rewardsTypes.QueryOutstandingRewardsRequest{
			RewardsAddress: "👻",
		})
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "invalid rewards address: decoding bech32 failed: invalid bech32 string length 4"), err)
	})

	s.Run("ok: get outstanding rewards", func() {
		res, err := querySrvr.OutstandingRewards(sdk.WrapSDKContext(ctx), &rewardsTypes.QueryOutstandingRewardsRequest{
			RewardsAddress: s.chain.GetAccount(0).Address.String(),
		})
		s.Require().NoError(err)
		s.Require().EqualValues(0, res.RecordsNum)
	})
}

func (s *KeeperTestSuite) TestGRPC_RewardsRecords() {
	ctx, k := s.chain.GetContext(), s.chain.GetApp().RewardsKeeper

	querySrvr := keeper.NewQueryServer(k)

	s.Run("err: empty request", func() {
		_, err := querySrvr.RewardsRecords(sdk.WrapSDKContext(ctx), nil)
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "empty request"), err)
	})

	s.Run("err: invalid rewards address", func() {
		_, err := querySrvr.RewardsRecords(sdk.WrapSDKContext(ctx), &rewardsTypes.QueryRewardsRecordsRequest{
			RewardsAddress: "👻",
		})
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "invalid rewards address: decoding bech32 failed: invalid bech32 string length 4"), err)
	})

	s.Run("ok: get rewards records", func() {
		res, err := querySrvr.RewardsRecords(sdk.WrapSDKContext(ctx), &rewardsTypes.QueryRewardsRecordsRequest{
			RewardsAddress: s.chain.GetAccount(0).Address.String(),
		})
		s.Require().NoError(err)
		s.Require().EqualValues(0, len(res.Records))
	})
}

func (s *KeeperTestSuite) TestGRPC_FlatFee() {
	ctx, k := s.chain.GetContext(), s.chain.GetApp().RewardsKeeper

	querySrvr := keeper.NewQueryServer(k)

	s.Run("err: empty request", func() {
		_, err := querySrvr.FlatFee(sdk.WrapSDKContext(ctx), nil)
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "empty request"), err)
	})

	s.Run("err: invalid contract address", func() {
		_, err := querySrvr.FlatFee(sdk.WrapSDKContext(ctx), &rewardsTypes.QueryFlatFeeRequest{
			ContractAddress: "👻",
		})
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "invalid contract address: decoding bech32 failed: invalid bech32 string length 4"), err)
	})

	s.Run("err: flat fee not found", func() {
		contractAddr := e2eTesting.GenContractAddresses(1)[0]
		_, err := querySrvr.FlatFee(sdk.WrapSDKContext(ctx), &rewardsTypes.QueryFlatFeeRequest{
			ContractAddress: contractAddr.String(),
		})
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.NotFound, "flat fee: not found"), err)
	})

	s.Run("ok: get flat fee", func() {
		contractAdminAcc := s.chain.GetAccount(0)
		contractViewer := testutils.NewMockContractViewer()
		k.SetContractInfoViewer(contractViewer)
		contractAddr := e2eTesting.GenContractAddresses(1)[0]
		contractViewer.AddContractAdmin(contractAddr.String(), contractAdminAcc.Address.String())
		err := k.SetContractMetadata(ctx, contractAdminAcc.Address, contractAddr, rewardsTypes.ContractMetadata{
			ContractAddress: contractAddr.String(),
			OwnerAddress:    contractAdminAcc.Address.String(),
		})
		s.Require().NoError(err)
		err = k.SetFlatFee(ctx, contractAdminAcc.Address, types.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         sdk.NewInt64Coin("token", 123),
		})
		s.Require().NoError(err)

		res, err := querySrvr.FlatFee(sdk.WrapSDKContext(ctx), &rewardsTypes.QueryFlatFeeRequest{
			ContractAddress: contractAddr.String(),
		})
		s.Require().NoError(err)
		s.Require().NotNil(res)
		s.Require().EqualValues(sdk.NewInt64Coin("token", 123), res.FlatFeeAmount)
	})
}
