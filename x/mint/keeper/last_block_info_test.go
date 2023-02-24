package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/archway-network/archway/x/mint/types"
)

func TestSetLastBlockInfo(t *testing.T) {
	currentTime := time.Now()
	testCases := []struct {
		testCase    string
		lbi         types.LastBlockInfo
		expectError bool
	}{
		{
			"invalid inflation amount",
			types.LastBlockInfo{
				Inflation: sdk.MustNewDecFromStr("123"),
			},
			true,
		},
		{
			"ok: valid inflation",
			types.LastBlockInfo{
				Inflation: sdk.MustNewDecFromStr("0.33"),
				Time:      &currentTime,
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testCase, func(t *testing.T) {
			keeper, ctx := SetupTestMintKeeper(t)

			err := keeper.SetLastBlockInfo(ctx, tc.lbi)

			if tc.expectError {
				require.Error(t, err, tc)
			} else {
				require.NoError(t, err, tc)
				lbi, _ := keeper.GetLastBlockInfo(ctx)
				require.EqualValues(t, tc.lbi.Inflation, lbi.Inflation, tc)
			}
		})
	}
}

func TestGetLastBlockInfo(t *testing.T) {
	keeper, ctx := SetupTestMintKeeper(t)
	currentTime := time.Now()

	// LastBlockInfo not found
	_, found := keeper.GetLastBlockInfo(ctx)
	require.False(t, found)

	// Save some block info
	lbi := types.LastBlockInfo{Inflation: sdk.MustNewDecFromStr("0.2"), Time: &currentTime}
	err := keeper.SetLastBlockInfo(ctx, lbi)
	require.NoError(t, err)
	res, found := keeper.GetLastBlockInfo(ctx)
	require.True(t, found)
	require.EqualValues(t, lbi.Inflation, res.Inflation)

	// Overwrite existing block info
	lbi2 := types.LastBlockInfo{Inflation: sdk.MustNewDecFromStr("0.3"), Time: &currentTime}
	err = keeper.SetLastBlockInfo(ctx, lbi2)
	require.NoError(t, err)
	res, found = keeper.GetLastBlockInfo(ctx)
	require.True(t, found)
	require.EqualValues(t, lbi2.Inflation, res.Inflation)
}
