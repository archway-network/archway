package keeper_test

import (
	"testing"

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

func TestGRPC_Params(t *testing.T) {
	k, ctx, _ := testutils.RewardsKeeper(t)
	querySrvr := keeper.NewQueryServer(k)
	params := rewardsTypes.Params{
		InflationRewardsRatio: math.LegacyMustNewDecFromStr("0.1"),
		TxFeeRebateRatio:      math.LegacyMustNewDecFromStr("0.1"),
		MaxWithdrawRecords:    uint64(2),
		MinPriceOfGas:         rewardsTypes.DefaultMinPriceOfGas,
	}
	err := k.Params.Set(ctx, params)
	require.NoError(t, err)

	t.Run("err: empty request", func(t *testing.T) {
		_, err := querySrvr.Params(ctx, nil)
		require.Error(t, err)
		require.Equal(t, status.Error(codes.InvalidArgument, "empty request"), err)
	})

	t.Run("ok: gets params", func(t *testing.T) {
		res, err := querySrvr.Params(ctx, &rewardsTypes.QueryParamsRequest{})
		require.NoError(t, err)
		require.Equal(t, params, res.Params)
	})
}

func TestGRPC_ContractMetadata(t *testing.T) {
	k, ctx, _ := testutils.RewardsKeeper(t)
	wk := testutils.NewMockContractViewer()
	k.SetContractInfoViewer(wk)
	querySrvr := keeper.NewQueryServer(k)

	contractAddr := e2eTesting.GenContractAddresses(2)
	contractAdminAcc := testutils.AccAddress()
	wk.AddContractAdmin(contractAddr[0].String(), contractAdminAcc.String())
	contractMeta := rewardsTypes.ContractMetadata{
		ContractAddress: contractAddr[0].String(),
		OwnerAddress:    contractAdminAcc.String(),
	}
	err := k.SetContractMetadata(ctx, contractAdminAcc, contractAddr[0], contractMeta)
	require.NoError(t, err)

	t.Run("err: empty request", func(t *testing.T) {
		_, err := querySrvr.ContractMetadata(ctx, nil)
		require.Error(t, err)
		require.Equal(t, status.Error(codes.InvalidArgument, "empty request"), err)
	})

	t.Run("err: invalid contract address", func(t *testing.T) {
		_, err := querySrvr.ContractMetadata(ctx, &rewardsTypes.QueryContractMetadataRequest{ContractAddress: "👻"})
		require.Error(t, err)
		require.Equal(t, status.Error(codes.InvalidArgument, "invalid contract address: decoding bech32 failed: invalid bech32 string length 4"), err)
	})

	t.Run("err: contract metadata not found", func(t *testing.T) {
		_, err := querySrvr.ContractMetadata(ctx, &rewardsTypes.QueryContractMetadataRequest{ContractAddress: contractAddr[1].String()})
		require.Error(t, err)
		require.Equal(t, status.Error(codes.NotFound, "metadata for the contract: not found"), err)
	})

	t.Run("ok: gets contract metadata", func(t *testing.T) {
		res, err := querySrvr.ContractMetadata(ctx, &rewardsTypes.QueryContractMetadataRequest{ContractAddress: contractAddr[0].String()})
		require.NoError(t, err)
		require.Equal(t, contractMeta.ContractAddress, res.Metadata.ContractAddress)
		require.Equal(t, contractMeta.OwnerAddress, res.Metadata.OwnerAddress)
		require.Equal(t, contractMeta.RewardsAddress, res.Metadata.RewardsAddress)
	})
}

func TestGRPC_BlockRewardsTracking(t *testing.T) {
	k, ctx, _ := testutils.RewardsKeeper(t)
	wk := testutils.NewMockContractViewer()
	k.SetContractInfoViewer(wk)
	querySrvr := keeper.NewQueryServer(k)

	t.Run("err: empty request", func(t *testing.T) {
		_, err := querySrvr.BlockRewardsTracking(ctx, nil)
		require.Error(t, err)
		require.Equal(t, status.Error(codes.InvalidArgument, "empty request"), err)
	})

	t.Run("ok: gets block rewards tracking", func(t *testing.T) {
		res, err := querySrvr.BlockRewardsTracking(ctx, &rewardsTypes.QueryBlockRewardsTrackingRequest{})
		require.NoError(t, err)
		require.Equal(t, 0, len(res.Block.TxRewards))
		require.Equal(t, ctx.BlockHeight(), res.Block.InflationRewards.Height)
	})
}

func TestGRPC_RewardsPool(t *testing.T) {
	k, ctx, _ := testutils.RewardsKeeper(t)
	wk := testutils.NewMockContractViewer()
	k.SetContractInfoViewer(wk)
	querySrvr := keeper.NewQueryServer(k)

	t.Run("err: empty request", func(t *testing.T) {
		_, err := querySrvr.RewardsPool(ctx, nil)
		require.Error(t, err)
		require.Equal(t, status.Error(codes.InvalidArgument, "empty request"), err)
	})

	t.Run("ok: gets rewards pool", func(t *testing.T) {
		res, err := querySrvr.RewardsPool(ctx, &rewardsTypes.QueryRewardsPoolRequest{})
		require.NoError(t, err)
		require.NotNil(t, res)
	})
}

func TestGRPC_EstimateTxFees(t *testing.T) {
	k, ctx, _ := testutils.RewardsKeeper(t)
	wk := testutils.NewMockContractViewer()
	k.SetContractInfoViewer(wk)
	querySrvr := keeper.NewQueryServer(k)

	t.Run("err: empty request", func(t *testing.T) {
		_, err := querySrvr.EstimateTxFees(ctx, nil)
		require.Error(t, err)
		require.Equal(t, status.Error(codes.InvalidArgument, "empty request"), err)
	})

	t.Run("ok: gets estimated tx fees", func(t *testing.T) {
		expectedFee := sdk.NewInt64Coin("stake", 0)
		res, err := querySrvr.EstimateTxFees(ctx, &rewardsTypes.QueryEstimateTxFeesRequest{GasLimit: 0})
		require.NoError(t, err)
		require.NotNil(t, res)
		fees := sdk.NewCoins(res.EstimatedFee...)
		require.EqualValues(t, expectedFee.Amount, fees.AmountOf("stake"))
	})

	minConsFee := sdk.NewInt64Coin("stake", 100)
	t.Run("ok: gets estimated tx fees (custom minconsfee set)", func(t *testing.T) {
		err := k.MinConsFee.Set(ctx, sdk.NewDecCoinFromCoin(minConsFee))
		require.NoError(t, err)
		res, err := querySrvr.EstimateTxFees(ctx, &rewardsTypes.QueryEstimateTxFeesRequest{GasLimit: 1})
		require.NoError(t, err)
		require.NotNil(t, res)
		fees := sdk.NewCoins(res.EstimatedFee...)
		require.EqualValues(t, minConsFee.Amount, fees.AmountOf("stake"))
	})

	t.Run("ok: gets estimated tx fees inclulding contract flat fee(diff denom)", func(t *testing.T) {
		expectedFlatFee := sdk.NewInt64Coin("token", 123)
		contractAdminAcc := testutils.AccAddress()
		contractAddr := e2eTesting.GenContractAddresses(1)[0]
		wk.AddContractAdmin(contractAddr.String(), contractAdminAcc.String())
		err := k.SetContractMetadata(ctx, contractAdminAcc, contractAddr, rewardsTypes.ContractMetadata{
			ContractAddress: contractAddr.String(),
			OwnerAddress:    contractAdminAcc.String(),
			RewardsAddress:  contractAdminAcc.String(),
		})
		require.NoError(t, err)
		err = k.SetFlatFee(ctx, contractAdminAcc, types.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         expectedFlatFee,
		})
		require.NoError(t, err)

		res, err := querySrvr.EstimateTxFees(ctx, &rewardsTypes.QueryEstimateTxFeesRequest{GasLimit: 1, ContractAddress: contractAddr.String()})
		require.NoError(t, err)
		require.NotNil(t, res)
		fees := sdk.NewCoins(res.EstimatedFee...)
		require.Equal(t, expectedFlatFee.Amount, fees.AmountOf("token"))
		require.EqualValues(t, minConsFee.Amount, fees.AmountOf("stake"))
	})

	t.Run("ok: gets estimated tx fees including contract flat fee(same denom)", func(t *testing.T) {
		expectedFlatFee := sdk.NewInt64Coin("stake", 123)
		contractAdminAcc := testutils.AccAddress()
		contractAddr := e2eTesting.GenContractAddresses(1)[0]
		wk.AddContractAdmin(contractAddr.String(), contractAdminAcc.String())
		err := k.SetContractMetadata(ctx, contractAdminAcc, contractAddr, rewardsTypes.ContractMetadata{
			ContractAddress: contractAddr.String(),
			OwnerAddress:    contractAdminAcc.String(),
			RewardsAddress:  contractAdminAcc.String(),
		})
		require.NoError(t, err)
		err = k.SetFlatFee(ctx, contractAdminAcc, types.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         expectedFlatFee,
		})
		require.NoError(t, err)

		res, err := querySrvr.EstimateTxFees(ctx, &rewardsTypes.QueryEstimateTxFeesRequest{GasLimit: 1, ContractAddress: contractAddr.String()})
		require.NoError(t, err)
		require.NotNil(t, res)
		fees := sdk.NewCoins(res.EstimatedFee...)
		require.Equal(t, expectedFlatFee.Add(minConsFee).Amount, fees.AmountOf("stake"))
	})
}

func TestGRPC_OutstandingRewards(t *testing.T) {
	k, ctx, _ := testutils.RewardsKeeper(t)
	wk := testutils.NewMockContractViewer()
	k.SetContractInfoViewer(wk)
	querySrvr := keeper.NewQueryServer(k)

	t.Run("err: empty request", func(t *testing.T) {
		_, err := querySrvr.OutstandingRewards(ctx, nil)
		require.Error(t, err)
		require.Equal(t, status.Error(codes.InvalidArgument, "empty request"), err)
	})

	t.Run("err: invalid rewards address", func(t *testing.T) {
		_, err := querySrvr.OutstandingRewards(ctx, &rewardsTypes.QueryOutstandingRewardsRequest{
			RewardsAddress: "👻",
		})
		require.Error(t, err)
		require.Equal(t, status.Error(codes.InvalidArgument, "invalid rewards address: decoding bech32 failed: invalid bech32 string length 4"), err)
	})

	t.Run("ok: get outstanding rewards", func(t *testing.T) {
		res, err := querySrvr.OutstandingRewards(ctx, &rewardsTypes.QueryOutstandingRewardsRequest{
			RewardsAddress: testutils.AccAddress().String(),
		})
		require.NoError(t, err)
		require.EqualValues(t, 0, res.RecordsNum)
	})
}

func TestGRPC_RewardsRecords(t *testing.T) {
	k, ctx, _ := testutils.RewardsKeeper(t)
	wk := testutils.NewMockContractViewer()
	k.SetContractInfoViewer(wk)
	querySrvr := keeper.NewQueryServer(k)

	t.Run("err: empty request", func(t *testing.T) {
		_, err := querySrvr.RewardsRecords(ctx, nil)
		require.Error(t, err)
		require.Equal(t, status.Error(codes.InvalidArgument, "empty request"), err)
	})

	t.Run("err: invalid rewards address", func(t *testing.T) {
		_, err := querySrvr.RewardsRecords(ctx, &rewardsTypes.QueryRewardsRecordsRequest{
			RewardsAddress: "👻",
		})
		require.Error(t, err)
		require.Equal(t, status.Error(codes.InvalidArgument, "invalid rewards address: decoding bech32 failed: invalid bech32 string length 4"), err)
	})

	t.Run("ok: get rewards records", func(t *testing.T) {
		res, err := querySrvr.RewardsRecords(ctx, &rewardsTypes.QueryRewardsRecordsRequest{
			RewardsAddress: testutils.AccAddress().String(),
		})
		require.NoError(t, err)
		require.EqualValues(t, 0, len(res.Records))
	})
}

func TestGRPC_FlatFee(t *testing.T) {
	k, ctx, _ := testutils.RewardsKeeper(t)
	wk := testutils.NewMockContractViewer()
	k.SetContractInfoViewer(wk)
	querySrvr := keeper.NewQueryServer(k)

	t.Run("err: empty request", func(t *testing.T) {
		_, err := querySrvr.FlatFee(ctx, nil)
		require.Error(t, err)
		require.Equal(t, status.Error(codes.InvalidArgument, "empty request"), err)
	})

	t.Run("err: invalid contract address", func(t *testing.T) {
		_, err := querySrvr.FlatFee(ctx, &rewardsTypes.QueryFlatFeeRequest{
			ContractAddress: "👻",
		})
		require.Error(t, err)
		require.Equal(t, status.Error(codes.InvalidArgument, "invalid contract address: decoding bech32 failed: invalid bech32 string length 4"), err)
	})

	t.Run("err: flat fee not found", func(t *testing.T) {
		contractAddr := e2eTesting.GenContractAddresses(1)[0]
		_, err := querySrvr.FlatFee(ctx, &rewardsTypes.QueryFlatFeeRequest{
			ContractAddress: contractAddr.String(),
		})
		require.Error(t, err)
		require.Equal(t, status.Error(codes.NotFound, "flat fee: not found"), err)
	})

	t.Run("ok: get flat fee", func(t *testing.T) {
		contractAdminAcc := testutils.AccAddress()
		contractAddr := e2eTesting.GenContractAddresses(1)[0]
		wk.AddContractAdmin(contractAddr.String(), contractAdminAcc.String())
		err := k.SetContractMetadata(ctx, contractAdminAcc, contractAddr, rewardsTypes.ContractMetadata{
			ContractAddress: contractAddr.String(),
			OwnerAddress:    contractAdminAcc.String(),
			RewardsAddress:  contractAdminAcc.String(),
		})
		require.NoError(t, err)
		err = k.SetFlatFee(ctx, contractAdminAcc, types.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         sdk.NewInt64Coin("token", 123),
		})
		require.NoError(t, err)

		res, err := querySrvr.FlatFee(ctx, &rewardsTypes.QueryFlatFeeRequest{
			ContractAddress: contractAddr.String(),
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, sdk.NewInt64Coin("token", 123), res.FlatFeeAmount)
	})
}
