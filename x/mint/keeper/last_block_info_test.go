package keeper_test

import (
	"testing"

	"github.com/archway-network/archway/x/mint/types"
	"github.com/stretchr/testify/require"
)

func TestSetLastBlockInfo(t *testing.T) {
	testCases := []struct {
		testCase    string
		lbi         types.LastBlockInfo
		expectError bool
	}{
		{
			"invalid inflation string",
			types.LastBlockInfo{
				Inflation: "👻",
			},
			true,
		},
		{
			"ok: valid inflation",
			types.LastBlockInfo{
				Inflation: "0.33",
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
				_, lbi := keeper.GetLastBlockInfo(ctx)
				require.EqualValues(t, tc.lbi, lbi, tc)
			}
		})
	}
}

func TestGetLastBlockInfo(t *testing.T) {
	keeper, ctx := SetupTestMintKeeper(t)

	// LastBlockInfo not found
	found, _ := keeper.GetLastBlockInfo(ctx)
	require.False(t, found)

	// Save some block info
	lbi := types.LastBlockInfo{Inflation: "0.2"}
	err := keeper.SetLastBlockInfo(ctx, lbi)
	require.NoError(t, err)
	found, res := keeper.GetLastBlockInfo(ctx)
	require.True(t, found)
	require.EqualValues(t, lbi, res)

	// Overwrite existing block info
	lbi2 := types.LastBlockInfo{Inflation: "0.3"}
	err = keeper.SetLastBlockInfo(ctx, lbi2)
	require.NoError(t, err)
	found, res = keeper.GetLastBlockInfo(ctx)
	require.True(t, found)
	require.EqualValues(t, lbi2, res)
}
