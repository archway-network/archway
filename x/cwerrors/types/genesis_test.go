package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/archway-network/archway/x/cwerrors/types"
)

func TestGenesisValidate(t *testing.T) {
	type testCase struct {
		name        string
		genesis     types.GenesisState
		errExpected bool
	}

	testCases := []testCase{
		{
			name:        "Fail: Empty values",
			genesis:     types.GenesisState{},
			errExpected: true,
		},
		{
			name: "Fail: Invalid params",
			genesis: types.GenesisState{
				Params: types.NewParams(
					0,
					true,
					sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
					100,
				),
			},
			errExpected: true,
		},
		{
			name: "OK: Valid genesis state",
			genesis: types.GenesisState{
				Params: types.DefaultParams(),
			},
			errExpected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.genesis.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
