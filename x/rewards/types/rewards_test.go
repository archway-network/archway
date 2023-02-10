package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
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
				TxId:   1,
				Height: 1,
				FeeRewards: []sdk.Coin{
					{Denom: "uarch", Amount: sdk.OneInt()},
					{},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid FeeRewards (invalid coin)",
			record: rewardsTypes.TxRewards{
				TxId:       1,
				Height:     1,
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

func TestRewardsRecordValidate(t *testing.T) {
	type testCase struct {
		name        string
		record      rewardsTypes.RewardsRecord
		errExpected bool
	}

	accAddrs, _ := e2eTesting.GenAccounts(1)
	accAddr := accAddrs[0]
	mockTime := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)

	testCases := []testCase{
		{
			name: "OK",
			record: rewardsTypes.RewardsRecord{
				Id:             1,
				RewardsAddress: accAddr.String(),
				Rewards: []sdk.Coin{
					{Denom: "uatom", Amount: sdk.OneInt()},
				},
				CalculatedHeight: 1,
				CalculatedTime:   mockTime,
			},
		},
		{
			name: "Fail: invalid Id",
			record: rewardsTypes.RewardsRecord{
				Id:             0,
				RewardsAddress: accAddr.String(),
				Rewards: []sdk.Coin{
					{Denom: "uatom", Amount: sdk.OneInt()},
				},
				CalculatedHeight: 1,
				CalculatedTime:   mockTime,
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid RewardsAddress",
			record: rewardsTypes.RewardsRecord{
				Id:             1,
				RewardsAddress: "invalid",
				Rewards: []sdk.Coin{
					{Denom: "uatom", Amount: sdk.OneInt()},
				},
				CalculatedHeight: 1,
				CalculatedTime:   mockTime,
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid Rewards (invalid coin)",
			record: rewardsTypes.RewardsRecord{
				Id:             1,
				RewardsAddress: accAddr.String(),
				Rewards: []sdk.Coin{
					{Denom: "uatom", Amount: sdk.NewInt(-1)},
				},
				CalculatedHeight: 1,
				CalculatedTime:   mockTime,
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid CalculatedHeight",
			record: rewardsTypes.RewardsRecord{
				Id:             1,
				RewardsAddress: accAddr.String(),
				Rewards: []sdk.Coin{
					{Denom: "uatom", Amount: sdk.OneInt()},
				},
				CalculatedHeight: -1,
				CalculatedTime:   mockTime,
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid CalculatedTime (empty)",
			record: rewardsTypes.RewardsRecord{
				Id:             1,
				RewardsAddress: accAddr.String(),
				Rewards: []sdk.Coin{
					{Denom: "uatom", Amount: sdk.OneInt()},
				},
				CalculatedHeight: 1,
				CalculatedTime:   time.Time{},
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

func TestFlatFeeValidate(t *testing.T) {
	type testCase struct {
		name        string
		flatFee     rewardsTypes.FlatFee
		errExpected bool
	}

	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	testCases := []testCase{
		{
			name: "OK: with flat fee coin",
			flatFee: rewardsTypes.FlatFee{
				ContractAddress: contractAddr.String(),
				FlatFee:         sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),
			},
		},
		{
			name: "Fail: invalid ContractAddress",
			flatFee: rewardsTypes.FlatFee{
				ContractAddress: "invalid",
			},
			errExpected: true,
		},
		{
			name: "Fail: empty flat fee",
			flatFee: rewardsTypes.FlatFee{
				ContractAddress: contractAddr.String(),
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid flat fee amount",
			flatFee: rewardsTypes.FlatFee{
				ContractAddress: contractAddr.String(),
				FlatFee: sdk.Coin{
					Denom:  sdk.DefaultBondDenom,
					Amount: sdk.NewInt(-1),
				},
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.flatFee.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
