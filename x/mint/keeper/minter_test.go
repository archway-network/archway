package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/archway-network/archway/x/mint/types"
)

func TestGetBlockProvisions(t *testing.T) {
	currentTime := time.Now()
	hourAgo := currentTime.Add(-time.Hour * 1)
	testCases := []struct {
		testCase        string
		lbi             types.LastBlockInfo
		expectTokens    sdk.Dec
		expectInflation sdk.Dec
	}{
		{
			"ok: just minted. should not mint more tokens",
			types.LastBlockInfo{
				Inflation: sdk.MustNewDecFromStr("0.33"),
				Time:      &currentTime,
			},
			sdk.ZeroDec(),
			sdk.MustNewDecFromStr("0.33"),
		},
		{
			"ok: last mint was an hour ago",
			types.LastBlockInfo{
				Inflation: sdk.MustNewDecFromStr("0.33"),
				Time:      &hourAgo,
			},
			sdk.MustNewDecFromStr("0.000031392694063912"),
			sdk.MustNewDecFromStr("0.33"),
		},
		{
			"ok: last mint was an hour ago. but inflation is 10%",
			types.LastBlockInfo{
				Inflation: sdk.MustNewDecFromStr("0.1"),
				Time:      &hourAgo,
			},
			sdk.MustNewDecFromStr("0.000009512937595125"),
			sdk.MustNewDecFromStr("0.1"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testCase, func(t *testing.T) {
			keeper, ctx := SetupTestMintKeeper(t)
			err := keeper.SetLastBlockInfo(ctx, tc.lbi)
			require.NoError(t, err, tc)

			tokens, inflation := keeper.GetBlockProvisions(ctx)

			require.EqualValues(t, tc.expectTokens, tokens)
			require.EqualValues(t, tc.expectInflation, inflation)
		})
	}
}
