package pkg

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitCoins(t *testing.T) {
	type testCase struct {
		coins          string
		ratio          string
		stack1Expected string
		stack2Expected string
	}

	testCases := []testCase{
		{
			coins:          "100uatom",
			ratio:          "0.5",
			stack1Expected: "50uatom",
			stack2Expected: "50uatom",
		},
		{
			coins:          "100uatom",
			ratio:          "0.75",
			stack1Expected: "75uatom",
			stack2Expected: "25uatom",
		},
		{
			coins:          "11uatom",
			ratio:          "0.50",
			stack1Expected: "5uatom",
			stack2Expected: "6uatom",
		},
		{
			coins:          "13uatom,20ubtc",
			ratio:          "0.25",
			stack1Expected: "3uatom,5ubtc",
			stack2Expected: "10uatom,15ubtc",
		},
		{
			coins:          "13uatom,20ubtc",
			ratio:          "1.0",
			stack1Expected: "13uatom,20ubtc",
			stack2Expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s -> {%s , %s} with %s", tc.coins, tc.stack1Expected, tc.stack2Expected, tc.ratio), func(t *testing.T) {
			coins, err := sdk.ParseCoinsNormalized(tc.coins)
			require.NoError(t, err)

			ratio, err := sdk.NewDecFromStr(tc.ratio)
			require.NoError(t, err)

			stack1Expected, err := sdk.ParseCoinsNormalized(tc.stack1Expected)
			require.NoError(t, err)

			stack2Expected, err := sdk.ParseCoinsNormalized(tc.stack2Expected)
			require.NoError(t, err)

			stack1Received, stack2Received := SplitCoins(coins, ratio)
			if tc.stack1Expected == "" {
				assert.True(t, stack1Received.Empty())
			}
			if tc.stack2Expected == "" {
				assert.True(t, stack2Received.Empty())
			}

			assert.ElementsMatch(t, stack1Expected, stack1Received)
			assert.ElementsMatch(t, stack2Expected, stack2Received)
		})
	}
}
