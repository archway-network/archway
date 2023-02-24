package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/archway-network/archway/x/mint/types"
)

func TestValidate(t *testing.T) {
	currentTime := time.Now()
	testCases := []struct {
		testCase    string
		lbi         types.LastBlockInfo
		expectError bool
	}{
		{
			"invalid inflation: less than 0: should be: 0 < inflation < 1",
			types.LastBlockInfo{
				Inflation: sdk.MustNewDecFromStr("1.0000001"),
			},
			true,
		},
		{
			"invalid inflation: more than 1: should be: 0 < inflation < 1",
			types.LastBlockInfo{
				Inflation: sdk.MustNewDecFromStr("-0.0000001"),
			},
			true,
		},
		{
			"invalid timestamp",
			types.LastBlockInfo{
				Inflation: sdk.MustNewDecFromStr("0.5"),
			},
			true,
		},
		{
			"ok: valid inflation",
			types.LastBlockInfo{
				Inflation: sdk.MustNewDecFromStr("0.5"),
				Time:      &currentTime,
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testCase, func(t *testing.T) {
			err := tc.lbi.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
