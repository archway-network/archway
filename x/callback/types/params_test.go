package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/archway-network/archway/x/callback/types"
	"github.com/stretchr/testify/assert"
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
				100,
				100,
				math.LegacyMustNewDecFromStr("1.0"),
				math.LegacyMustNewDecFromStr("1.0"),
			),
			errExpected: false,
		},
		{
			name: "Fail: CallbackGasLimit: zero",
			params: types.NewParams(
				0,
				100,
				100,
				math.LegacyMustNewDecFromStr("1.0"),
				math.LegacyMustNewDecFromStr("1.0"),
			),
			errExpected: true,
		},
		{
			name: "Fail: BlockReservationFeeMultiplier: negative",
			params: types.NewParams(
				100,
				100,
				100,
				math.LegacyMustNewDecFromStr("-1.0"),
				math.LegacyMustNewDecFromStr("1.0"),
			),
			errExpected: true,
		},
		{
			name: "Fail: FutureReservationFeeMultiplier: negative",
			params: types.NewParams(
				100,
				100,
				100,
				math.LegacyMustNewDecFromStr("1.0"),
				math.LegacyMustNewDecFromStr("-1.0"),
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
