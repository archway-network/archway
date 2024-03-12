package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/archway-network/archway/x/cwerrors/types"
)

func TestParamsValidate(t *testing.T) {
	type testCase struct {
		name        string
		params      types.Params
		errExpected bool
	}

	testCases := []testCase{
		{
			name:        "OK: Default values",
			params:      types.DefaultParams(),
			errExpected: false,
		},
		{
			name: "OK: All valid values",
			params: types.NewParams(
				100,
				sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
				100,
			),
			errExpected: false,
		},
		{
			name: "Fail: ErrorStoredTime: zero",
			params: types.NewParams(
				0,
				sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
				100,
			),
			errExpected: true,
		},
		{
			name: "Fail: ErrorStoredTime: negative",
			params: types.NewParams(
				-2,
				sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
				100,
			),
			errExpected: true,
		},
		{
			name: "Fail: SubsciptionFee: invalid",
			params: types.NewParams(
				100,
				sdk.Coin{Denom: "", Amount: sdk.NewInt(100)},
				100,
			),
			errExpected: true,
		},
		{
			name: "Fail: SubscriptionPeriod: zero",
			params: types.NewParams(
				100,
				sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
				-2,
			),
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
