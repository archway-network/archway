package types_test

import (
	"testing"

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
				true,
			),
			errExpected: false,
		},
		{
			name: "Fail: ErrorStoredTime: zero",
			params: types.NewParams(
				0,
				true,
			),
			errExpected: true,
		},
		{
			name: "Fail: ErrorStoredTime: negative",
			params: types.NewParams(
				-2,
				true,
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
