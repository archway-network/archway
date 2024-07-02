package types_test

import (
	"testing"

	math "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

func TestRewardsParamsValidate(t *testing.T) {
	type testCase struct {
		name        string
		params      rewardsTypes.Params
		errExpected bool
	}

	testCases := []testCase{
		{
			name: "OK",
			params: rewardsTypes.Params{
				InflationRewardsRatio: math.LegacyNewDecWithPrec(2, 2),
				TxFeeRebateRatio:      math.LegacyNewDecWithPrec(5, 2),
				MaxWithdrawRecords:    1,
				MinPriceOfGas:         rewardsTypes.DefaultMinPriceOfGas,
			},
		},
		{
			name: "Fail: InflationRewardsRatio: negative",
			params: rewardsTypes.Params{
				InflationRewardsRatio: math.LegacyNewDecWithPrec(-2, 2),
				TxFeeRebateRatio:      math.LegacyNewDecWithPrec(5, 2),
				MaxWithdrawRecords:    1,
				MinPriceOfGas:         rewardsTypes.DefaultMinPriceOfGas,
			},
			errExpected: true,
		},
		{
			name: "Fail: InflationRewardsRatio: equal to 1.0",
			params: rewardsTypes.Params{
				InflationRewardsRatio: math.LegacyNewDecWithPrec(1, 0),
				TxFeeRebateRatio:      math.LegacyNewDecWithPrec(5, 2),
				MaxWithdrawRecords:    1,
				MinPriceOfGas:         rewardsTypes.DefaultMinPriceOfGas,
			},
			errExpected: true,
		},
		{
			name: "Fail: TxFeeRebateRatio: negative",
			params: rewardsTypes.Params{
				InflationRewardsRatio: math.LegacyNewDecWithPrec(2, 2),
				TxFeeRebateRatio:      math.LegacyNewDecWithPrec(-1, 2),
				MaxWithdrawRecords:    1,
				MinPriceOfGas:         rewardsTypes.DefaultMinPriceOfGas,
			},
			errExpected: true,
		},
		{
			name: "Fail: TxFeeRebateRatio: equal to 1.0",
			params: rewardsTypes.Params{
				InflationRewardsRatio: math.LegacyNewDecWithPrec(2, 2),
				TxFeeRebateRatio:      math.LegacyNewDecWithPrec(1, 0),
				MaxWithdrawRecords:    1,
				MinPriceOfGas:         rewardsTypes.DefaultMinPriceOfGas,
			},
			errExpected: true,
		},
		{
			name: "Fail: MaxWithdrawRecords: empty",
			params: rewardsTypes.Params{
				InflationRewardsRatio: math.LegacyNewDecWithPrec(2, 2),
				TxFeeRebateRatio:      math.LegacyNewDecWithPrec(1, 0),
				MaxWithdrawRecords:    0,
				MinPriceOfGas:         rewardsTypes.DefaultMinPriceOfGas,
			},
			errExpected: true,
		},
		{
			name: "Fail: MinPriceOfGas: empty",
			params: rewardsTypes.Params{
				InflationRewardsRatio: math.LegacyNewDecWithPrec(2, 2),
				TxFeeRebateRatio:      math.LegacyNewDecWithPrec(1, 0),
				MaxWithdrawRecords:    1000,
			},
			errExpected: true,
		},
		{
			name: "Fail: MinPriceOfGas: negative",
			params: rewardsTypes.Params{
				InflationRewardsRatio: math.LegacyNewDecWithPrec(2, 2),
				TxFeeRebateRatio:      math.LegacyNewDecWithPrec(1, 0),
				MaxWithdrawRecords:    1000,
				MinPriceOfGas:         sdk.DecCoin{Denom: "stake", Amount: math.LegacyNewDec(-100)},
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
