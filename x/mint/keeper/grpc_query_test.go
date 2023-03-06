package keeper_test

import (
	"testing"
	"time"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/archway-network/archway/x/mint/keeper"
	"github.com/archway-network/archway/x/mint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGRPC_Params(t *testing.T) {
	k, ctx := SetupTestMintKeeper(t)
	queryServer := keeper.NewQueryServer(k)
	params := types.Params{
		MinInflation:     sdk.MustNewDecFromStr("0.2"),
		MaxInflation:     sdk.MustNewDecFromStr("0.2"),
		MinBonded:        sdk.MustNewDecFromStr("0.2"),
		MaxBonded:        sdk.MustNewDecFromStr("0.2"),
		InflationChange:  sdk.MustNewDecFromStr("0.2"),
		MaxBlockDuration: time.Hour,
		InflationRecipients: []*types.InflationRecipient{
			{
				Recipient: types.ModuleName,
				Ratio:     sdk.MustNewDecFromStr("0.2"),
			},
			{
				Recipient: authtypes.FeeCollectorName,
				Ratio:     sdk.MustNewDecFromStr("0.8"),
			},
		},
	}
	k.SetParams(ctx, params)

	t.Run("err: empty request", func(t *testing.T) {
		_, err := queryServer.Params(sdk.WrapSDKContext(ctx), nil)
		assert.Error(t, err)
		assert.Equal(t, status.Error(codes.InvalidArgument, "empty request"), err)
	})

	t.Run("ok: gets params", func(t *testing.T) {
		res, err := queryServer.Params(sdk.WrapSDKContext(ctx), &types.QueryParamsRequest{})
		assert.NoError(t, err)
		assert.Equal(t, params.MinInflation, res.Params.MinInflation)
		assert.Equal(t, params.MaxInflation, res.Params.MaxInflation)
		assert.Equal(t, params.MinBonded, res.Params.MinBonded)
		assert.Equal(t, params.MaxBonded, res.Params.MaxBonded)
		assert.Equal(t, params.InflationChange, res.Params.InflationChange)
		assert.Equal(t, params.MaxBlockDuration, res.Params.MaxBlockDuration)
		assert.Equal(t, params.InflationRecipients, res.Params.InflationRecipients)
	})
}

func TestGRPC_Inflation(t *testing.T) {
	k, ctx := SetupTestMintKeeper(t)
	queryServer := keeper.NewQueryServer(k)
	now := time.Now()
	lastBlockInfo := types.LastBlockInfo{
		Inflation: sdk.MustNewDecFromStr("0.2"),
		Time:      &now,
	}

	t.Run("err: empty request", func(t *testing.T) {
		_, err := queryServer.Inflation(sdk.WrapSDKContext(ctx), nil)
		assert.Error(t, err)
		assert.Equal(t, status.Error(codes.InvalidArgument, "empty request"), err)
	})

	t.Run("err: last block info not found", func(t *testing.T) {
		_, err := queryServer.Inflation(sdk.WrapSDKContext(ctx), &types.QueryInflationRequest{})
		assert.Error(t, err)
		assert.Equal(t, status.Error(codes.NotFound, "inflation data not found"), err)
	})

	err := k.SetLastBlockInfo(ctx, lastBlockInfo)
	assert.NoError(t, err)

	t.Run("ok: gets inflation", func(t *testing.T) {
		res, err := queryServer.Inflation(sdk.WrapSDKContext(ctx), &types.QueryInflationRequest{})
		assert.NoError(t, err)
		assert.Equal(t, lastBlockInfo.Inflation, res.Inflation)
	})
}
