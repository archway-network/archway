package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

func TestBlockRewardsValidate(t *testing.T) {
	type testCase struct {
		name        string
		record      rewardsTypes.BlockRewards
		errExpected bool
	}

	testCases := []testCase{
		{
			name: "OK",
			record: rewardsTypes.BlockRewards{
				Height:           1,
				InflationRewards: sdk.Coin{Denom: "uatom", Amount: sdk.OneInt()},
			},
		},
		{
			name: "OK: no rewards",
			record: rewardsTypes.BlockRewards{
				Height: 1,
			},
		},
		{
			name: "Fail: invalid Height",
			record: rewardsTypes.BlockRewards{
				Height:           -1,
				InflationRewards: sdk.Coin{Denom: "uatom", Amount: sdk.OneInt()},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid InflationRewards denom",
			record: rewardsTypes.BlockRewards{
				Height:           1,
				InflationRewards: sdk.Coin{Denom: "123invalid", Amount: sdk.OneInt()},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid InflationRewards amount (negative)",
			record: rewardsTypes.BlockRewards{
				Height:           1,
				InflationRewards: sdk.Coin{Denom: "uatom", Amount: sdk.NewInt(-1)},
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.record.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestTxRewardsValidate(t *testing.T) {
	type testCase struct {
		name        string
		record      rewardsTypes.TxRewards
		errExpected bool
	}

	testCases := []testCase{
		{
			name: "OK",
			record: rewardsTypes.TxRewards{
				TxId:       1,
				Height:     1,
				FeeRewards: sdk.NewCoins(sdk.NewCoin("uatom", sdk.OneInt())),
			},
		},
		{
			name: "OK: no rewards",
			record: rewardsTypes.TxRewards{
				TxId:   1,
				Height: 1,
			},
		},
		{
			name: "Fail: invalid TxId",
			record: rewardsTypes.TxRewards{
				TxId:       0,
				Height:     1,
				FeeRewards: sdk.NewCoins(sdk.NewCoin("uatom", sdk.OneInt())),
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid Height",
			record: rewardsTypes.TxRewards{
				TxId:       1,
				Height:     -1,
				FeeRewards: sdk.NewCoins(sdk.NewCoin("uatom", sdk.OneInt())),
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid FeeRewards (empty coin)",
			record: rewardsTypes.TxRewards{
				TxId:       1,
				Height:     -1,
				FeeRewards: sdk.NewCoins(),
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid FeeRewards (invalid coin)",
			record: rewardsTypes.TxRewards{
				TxId:       1,
				Height:     -1,
				FeeRewards: []sdk.Coin{{Denom: "123invalid", Amount: sdk.OneInt()}},
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.record.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
