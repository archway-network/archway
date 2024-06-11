package keeper_test

import (
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	math "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/rewards/keeper"
	"github.com/archway-network/archway/x/rewards/types"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

func (s *KeeperTestSuite) TestGRPC_Params() {
	querySrvr := keeper.NewQueryServer(s.keeper)
	params := rewardsTypes.Params{
		InflationRewardsRatio: math.LegacyMustNewDecFromStr("0.1"),
		TxFeeRebateRatio:      math.LegacyMustNewDecFromStr("0.1"),
		MaxWithdrawRecords:    uint64(2),
		MinPriceOfGas:         rewardsTypes.DefaultMinPriceOfGas,
	}
	err := s.keeper.Params.Set(s.ctx, params)
	require.NoError(s.T(), err)

	s.Run("err: empty request", func() {
		_, err := querySrvr.Params(s.ctx, nil)
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "empty request"), err)
	})

	s.Run("ok: gets params", func() {
		res, err := querySrvr.Params(s.ctx, &rewardsTypes.QueryParamsRequest{})
		s.Require().NoError(err)
		s.Require().Equal(params, res.Params)
	})
}

func (s *KeeperTestSuite) TestGRPC_ContractMetadata() {
	querySrvr := keeper.NewQueryServer(s.keeper)
	contractViewer := testutils.NewMockContractViewer()
	s.keeper.SetContractInfoViewer(contractViewer)
	contractAddr := e2eTesting.GenContractAddresses(2)
	contractAdminAcc := testutils.AccAddress()
	contractViewer.AddContractAdmin(contractAddr[0].String(), contractAdminAcc.String())
	contractMeta := rewardsTypes.ContractMetadata{
		ContractAddress: contractAddr[0].String(),
		OwnerAddress:    contractAdminAcc.String(),
	}
	err := s.keeper.SetContractMetadata(s.ctx, contractAdminAcc, contractAddr[0], contractMeta)
	s.Require().NoError(err)

	s.Run("err: empty request", func() {
		_, err := querySrvr.ContractMetadata(s.ctx, nil)
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "empty request"), err)
	})

	s.Run("err: invalid contract address", func() {
		_, err := querySrvr.ContractMetadata(s.ctx, &rewardsTypes.QueryContractMetadataRequest{ContractAddress: "👻"})
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "invalid contract address: decoding bech32 failed: invalid bech32 string length 4"), err)
	})

	s.Run("err: contract metadata not found", func() {
		_, err := querySrvr.ContractMetadata(s.ctx, &rewardsTypes.QueryContractMetadataRequest{ContractAddress: contractAddr[1].String()})
		s.Require().Error(err)
		s.Require().Equal(status.Errorf(codes.NotFound, "metadata for the contract: not found"), err)
	})

	s.Run("ok: gets contract metadata", func() {
		res, err := querySrvr.ContractMetadata(s.ctx, &rewardsTypes.QueryContractMetadataRequest{ContractAddress: contractAddr[0].String()})
		s.Require().NoError(err)
		s.Require().Equal(contractMeta.ContractAddress, res.Metadata.ContractAddress)
		s.Require().Equal(contractMeta.RewardsAddress, res.Metadata.RewardsAddress)
		s.Require().Equal(contractMeta.OwnerAddress, res.Metadata.OwnerAddress)
	})
}

func (s *KeeperTestSuite) TestGRPC_BlockRewardsTracking() {
	querySrvr := keeper.NewQueryServer(s.keeper)

	s.Run("err: empty request", func() {
		_, err := querySrvr.BlockRewardsTracking(s.ctx, nil)
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "empty request"), err)
	})

	s.Run("ok: gets block rewards tracking", func() {
		res, err := querySrvr.BlockRewardsTracking(s.ctx, &rewardsTypes.QueryBlockRewardsTrackingRequest{})
		s.Require().NoError(err)
		s.Require().Equal(0, len(res.Block.TxRewards))
		s.Require().Equal(s.ctx.BlockHeight(), res.Block.InflationRewards.Height)
	})
}

func (s *KeeperTestSuite) TestGRPC_RewardsPool() {
	querySrvr := keeper.NewQueryServer(s.keeper)

	s.Run("err: empty request", func() {
		_, err := querySrvr.RewardsPool(s.ctx, nil)
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "empty request"), err)
	})

	s.Run("ok: gets rewards pool", func() {
		res, err := querySrvr.RewardsPool(s.ctx, &rewardsTypes.QueryRewardsPoolRequest{})
		s.Require().NoError(err)
		s.Require().NotNil(res)
	})
}

func (s *KeeperTestSuite) TestGRPC_EstimateTxFees() {
	querySrvr := keeper.NewQueryServer(s.keeper)

	s.Run("err: empty request", func() {
		_, err := querySrvr.EstimateTxFees(s.ctx, nil)
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "empty request"), err)
	})

	s.Run("ok: gets estimated tx fees", func() {
		expectedFee := sdk.NewInt64Coin("stake", 0)
		res, err := querySrvr.EstimateTxFees(s.ctx, &rewardsTypes.QueryEstimateTxFeesRequest{GasLimit: 0})
		s.Require().NoError(err)
		s.Require().NotNil(res)
		fees := sdk.NewCoins(res.EstimatedFee...)
		s.Require().EqualValues(expectedFee.Amount, fees.AmountOf("stake"))
	})

	minConsFee := sdk.NewInt64Coin("stake", 100)
	s.Run("ok: gets estimated tx fees (custom minconsfee set)", func() {
		err := s.keeper.MinConsFee.Set(s.ctx, sdk.NewDecCoinFromCoin(minConsFee))
		s.Require().NoError(err)
		res, err := querySrvr.EstimateTxFees(s.ctx, &rewardsTypes.QueryEstimateTxFeesRequest{GasLimit: 1})
		s.Require().NoError(err)
		s.Require().NotNil(res)
		fees := sdk.NewCoins(res.EstimatedFee...)
		s.Require().EqualValues(minConsFee.Amount, fees.AmountOf("stake"))
	})

	s.Run("ok: gets estimated tx fees inclulding contract flat fee(diff denom)", func() {
		expectedFlatFee := sdk.NewInt64Coin("token", 123)
		contractAdminAcc := testutils.AccAddress()
		contractAddr := e2eTesting.GenContractAddresses(1)[0]
		s.wasmKeeper.AddContractAdmin(contractAddr.String(), contractAdminAcc.String())
		err := s.keeper.SetContractMetadata(s.ctx, contractAdminAcc, contractAddr, rewardsTypes.ContractMetadata{
			ContractAddress: contractAddr.String(),
			OwnerAddress:    contractAdminAcc.String(),
			RewardsAddress:  contractAdminAcc.String(),
		})
		s.Require().NoError(err)
		err = s.keeper.SetFlatFee(s.ctx, contractAdminAcc, types.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         expectedFlatFee,
		})
		s.Require().NoError(err)

		res, err := querySrvr.EstimateTxFees(s.ctx, &rewardsTypes.QueryEstimateTxFeesRequest{GasLimit: 1, ContractAddress: contractAddr.String()})
		s.Require().NoError(err)
		s.Require().NotNil(res)
		fees := sdk.NewCoins(res.EstimatedFee...)
		s.Require().Equal(expectedFlatFee.Amount, fees.AmountOf("token"))
		s.Require().EqualValues(minConsFee.Amount, fees.AmountOf("stake"))
	})

	s.Run("ok: gets estimated tx fees including contract flat fee(same denom)", func() {
		expectedFlatFee := sdk.NewInt64Coin("stake", 123)
		contractAdminAcc := testutils.AccAddress()
		contractAddr := e2eTesting.GenContractAddresses(1)[0]
		s.wasmKeeper.AddContractAdmin(contractAddr.String(), contractAdminAcc.String())
		err := s.keeper.SetContractMetadata(s.ctx, contractAdminAcc, contractAddr, rewardsTypes.ContractMetadata{
			ContractAddress: contractAddr.String(),
			OwnerAddress:    contractAdminAcc.String(),
			RewardsAddress:  contractAdminAcc.String(),
		})
		s.Require().NoError(err)
		err = s.keeper.SetFlatFee(s.ctx, contractAdminAcc, types.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         expectedFlatFee,
		})
		s.Require().NoError(err)

		res, err := querySrvr.EstimateTxFees(s.ctx, &rewardsTypes.QueryEstimateTxFeesRequest{GasLimit: 1, ContractAddress: contractAddr.String()})
		s.Require().NoError(err)
		s.Require().NotNil(res)
		fees := sdk.NewCoins(res.EstimatedFee...)
		s.Require().Equal(expectedFlatFee.Add(minConsFee).Amount, fees.AmountOf("stake"))
	})
}

func (s *KeeperTestSuite) TestGRPC_OutstandingRewards() {
	querySrvr := keeper.NewQueryServer(s.keeper)

	s.Run("err: empty request", func() {
		_, err := querySrvr.OutstandingRewards(s.ctx, nil)
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "empty request"), err)
	})

	s.Run("err: invalid rewards address", func() {
		_, err := querySrvr.OutstandingRewards(s.ctx, &rewardsTypes.QueryOutstandingRewardsRequest{
			RewardsAddress: "👻",
		})
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "invalid rewards address: decoding bech32 failed: invalid bech32 string length 4"), err)
	})

	s.Run("ok: get outstanding rewards", func() {
		res, err := querySrvr.OutstandingRewards(s.ctx, &rewardsTypes.QueryOutstandingRewardsRequest{
			RewardsAddress: testutils.AccAddress().String(),
		})
		s.Require().NoError(err)
		s.Require().EqualValues(0, res.RecordsNum)
	})
}

func (s *KeeperTestSuite) TestGRPC_RewardsRecords() {
	querySrvr := keeper.NewQueryServer(s.keeper)

	s.Run("err: empty request", func() {
		_, err := querySrvr.RewardsRecords(s.ctx, nil)
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "empty request"), err)
	})

	s.Run("err: invalid rewards address", func() {
		_, err := querySrvr.RewardsRecords(s.ctx, &rewardsTypes.QueryRewardsRecordsRequest{
			RewardsAddress: "👻",
		})
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "invalid rewards address: decoding bech32 failed: invalid bech32 string length 4"), err)
	})

	s.Run("ok: get rewards records", func() {
		res, err := querySrvr.RewardsRecords(s.ctx, &rewardsTypes.QueryRewardsRecordsRequest{
			RewardsAddress: testutils.AccAddress().String(),
		})
		s.Require().NoError(err)
		s.Require().EqualValues(0, len(res.Records))
	})
}

func (s *KeeperTestSuite) TestGRPC_FlatFee() {
	querySrvr := keeper.NewQueryServer(s.keeper)

	s.Run("err: empty request", func() {
		_, err := querySrvr.FlatFee(s.ctx, nil)
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "empty request"), err)
	})

	s.Run("err: invalid contract address", func() {
		_, err := querySrvr.FlatFee(s.ctx, &rewardsTypes.QueryFlatFeeRequest{
			ContractAddress: "👻",
		})
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.InvalidArgument, "invalid contract address: decoding bech32 failed: invalid bech32 string length 4"), err)
	})

	s.Run("err: flat fee not found", func() {
		contractAddr := e2eTesting.GenContractAddresses(1)[0]
		_, err := querySrvr.FlatFee(s.ctx, &rewardsTypes.QueryFlatFeeRequest{
			ContractAddress: contractAddr.String(),
		})
		s.Require().Error(err)
		s.Require().Equal(status.Error(codes.NotFound, "flat fee: not found"), err)
	})

	s.Run("ok: get flat fee", func() {
		contractAdminAcc := testutils.AccAddress()
		contractAddr := e2eTesting.GenContractAddresses(1)[0]
		s.wasmKeeper.AddContractAdmin(contractAddr.String(), contractAdminAcc.String())
		err := s.keeper.SetContractMetadata(s.ctx, contractAdminAcc, contractAddr, rewardsTypes.ContractMetadata{
			ContractAddress: contractAddr.String(),
			OwnerAddress:    contractAdminAcc.String(),
			RewardsAddress:  contractAdminAcc.String(),
		})
		s.Require().NoError(err)
		err = s.keeper.SetFlatFee(s.ctx, contractAdminAcc, types.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         sdk.NewInt64Coin("token", 123),
		})
		s.Require().NoError(err)

		res, err := querySrvr.FlatFee(s.ctx, &rewardsTypes.QueryFlatFeeRequest{
			ContractAddress: contractAddr.String(),
		})
		s.Require().NoError(err)
		s.Require().NotNil(res)
		s.Require().EqualValues(sdk.NewInt64Coin("token", 123), res.FlatFeeAmount)
	})
}
