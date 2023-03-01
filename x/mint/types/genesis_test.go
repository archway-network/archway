package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/archway-network/archway/x/mint/types"
)

func TestGenesisValidate(t *testing.T) {
	currentTime := time.Now()
	testCases := []struct {
		testCase    string
		gs          types.GenesisState
		expectError bool
	}{
		{
			"fail: invalid params",
			types.GenesisState{
				types.Params{
					MinInflation:     sdk.MustNewDecFromStr("0.2"),
					MaxInflation:     sdk.MustNewDecFromStr("0.2"),
					MinBonded:        sdk.MustNewDecFromStr("0.2"),
					MaxBonded:        sdk.MustNewDecFromStr("0.2"),
					InflationChange:  sdk.MustNewDecFromStr("0.2"),
					MaxBlockDuration: time.Hour,
				},
				types.LastBlockInfo{
					Inflation: sdk.MustNewDecFromStr("0.5"),
					Time:      &currentTime,
				},
			},
			true,
		},
		{
			"fail: invalid last block info",
			types.GenesisState{
				types.Params{
					MinInflation:     sdk.MustNewDecFromStr("0.2"),
					MaxInflation:     sdk.MustNewDecFromStr("0.2"),
					MinBonded:        sdk.MustNewDecFromStr("0.2"),
					MaxBonded:        sdk.MustNewDecFromStr("0.2"),
					InflationChange:  sdk.MustNewDecFromStr("0.2"),
					MaxBlockDuration: time.Hour,
					InflationRecipients: []*types.InflationRecipient{
						{
							Recipient: types.ModuleName,
							Ratio:     sdk.MustNewDecFromStr("1"),
						},
					},
				},
				types.LastBlockInfo{
					Inflation: sdk.MustNewDecFromStr("0.5"),
				},
			},
			true,
		},
		{
			"ok: all valid",
			types.GenesisState{
				types.Params{
					MinInflation:     sdk.MustNewDecFromStr("0.2"),
					MaxInflation:     sdk.MustNewDecFromStr("0.2"),
					MinBonded:        sdk.MustNewDecFromStr("0.2"),
					MaxBonded:        sdk.MustNewDecFromStr("0.2"),
					InflationChange:  sdk.MustNewDecFromStr("0.2"),
					MaxBlockDuration: time.Hour,
					InflationRecipients: []*types.InflationRecipient{
						{
							Recipient: types.ModuleName,
							Ratio:     sdk.MustNewDecFromStr("1"),
						},
					},
				},
				types.LastBlockInfo{
					Inflation: sdk.MustNewDecFromStr("0.5"),
					Time:      &currentTime,
				},
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testCase, func(t *testing.T) {
			err := tc.gs.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
