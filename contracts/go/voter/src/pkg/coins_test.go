package pkg

import (
	"testing"

	"github.com/CosmWasm/cosmwasm-go/std/math"
	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoinsContainMinAmount(t *testing.T) {
	type testCase struct {
		name  string
		coins []stdTypes.Coin
		coin  stdTypes.Coin
		//
		errExpected bool
	}

	testCases := []testCase{
		{
			name: "OK: GT",
			coins: []stdTypes.Coin{
				stdTypes.NewCoinFromUint64(100, "uatom"),
				stdTypes.NewCoinFromUint64(50, "musdt"),
			},
			coin: stdTypes.NewCoinFromUint64(45, "musdt"),
		},
		{
			name: "OK: GT",
			coins: []stdTypes.Coin{
				stdTypes.NewCoinFromUint64(100, "uatom"),
				stdTypes.NewCoinFromUint64(50, "musdt"),
			},
			coin: stdTypes.NewCoinFromUint64(50, "musdt"),
		},
		{
			name: "Fail: not found",
			coins: []stdTypes.Coin{
				stdTypes.NewCoinFromUint64(100, "uatom"),
				stdTypes.NewCoinFromUint64(50, "musdt"),
			},
			coin:        stdTypes.NewCoinFromUint64(10, "uusdc"),
			errExpected: true,
		},
		{
			name: "Fail: LT",
			coins: []stdTypes.Coin{
				stdTypes.NewCoinFromUint64(100, "uatom"),
				stdTypes.NewCoinFromUint64(50, "musdt"),
			},
			coin:        stdTypes.NewCoinFromUint64(55, "musdt"),
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.errExpected {
				assert.Error(t, CoinsContainMinAmount(tc.coins, tc.coin))
				return
			}
			assert.NoError(t, CoinsContainMinAmount(tc.coins, tc.coin))
		})
	}
}

func TestParseCoinFromString(t *testing.T) {
	type testCase struct {
		coinStr string
		//
		errExpected  bool
		coinExpected stdTypes.Coin
	}

	testCases := []testCase{
		{
			coinStr:      "1uatom",
			coinExpected: stdTypes.Coin{Denom: "uatom", Amount: math.NewUint128FromUint64(1)},
		},
		{
			coinStr:      "0ibc/312F13C9A9ECCE611FE8112B5ABCF0A14DE2C3937E38DEBF6B73F2534A83464E",
			coinExpected: stdTypes.Coin{Denom: "ibc/312F13C9A9ECCE611FE8112B5ABCF0A14DE2C3937E38DEBF6B73F2534A83464E", Amount: math.ZeroUint128()},
		},
		{
			coinStr:     "",
			errExpected: true,
		},
		{
			coinStr:     "uatom",
			errExpected: true,
		},
		{
			coinStr:     "123",
			errExpected: true,
		},
		{
			coinStr:     "123uatom#",
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.coinStr, func(t *testing.T) {
			coin, err := ParseCoinFromString(tc.coinStr)
			if tc.errExpected {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.coinExpected.Denom, coin.Denom)
			assert.Equal(t, tc.coinExpected.Amount.String(), coin.Amount.String())
		})
	}
}

func TestParseCoinsFromString(t *testing.T) {
	type testCase struct {
		coinsStr string
		//
		errExpected   bool
		coinsExpected []stdTypes.Coin
	}

	testCases := []testCase{
		{
			coinsStr: "1uatom",
			coinsExpected: []stdTypes.Coin{
				{Denom: "uatom", Amount: math.NewUint128FromUint64(1)},
			},
		},
		{
			coinsStr: "1uatom,2usdt",
			coinsExpected: []stdTypes.Coin{
				{Denom: "uatom", Amount: math.NewUint128FromUint64(1)},
				{Denom: "usdt", Amount: math.NewUint128FromUint64(2)},
			},
		},
		{
			coinsStr: "",
		},
		{
			coinsStr:    ",1uatom",
			errExpected: true,
		},
		{
			coinsStr:    "1musdt,,1uatom",
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.coinsStr, func(t *testing.T) {
			coins, err := ParseCoinsFromString(tc.coinsStr)
			if tc.errExpected {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.ElementsMatch(t, tc.coinsExpected, coins)
		})
	}
}
