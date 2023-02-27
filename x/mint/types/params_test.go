package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	"github.com/archway-network/archway/x/mint/types"
)

func TestParamsValidate(t *testing.T) {
	testCases := []struct {
		testCase    string
		params      types.Params
		expectError bool
	}{
		{
			"invalid minimum inflation: less than 0: should be: 0 < inflation < 1",
			types.Params{
				MinInflation: sdk.MustNewDecFromStr("2"),
			},
			true,
		},
		{
			"invalid maximum inflation: less than 0: should be: 0 < inflation < 1",
			types.Params{
				MinInflation: sdk.MustNewDecFromStr("0.2"),
				MaxInflation: sdk.MustNewDecFromStr("2"),
			},
			true,
		},
		{
			"invalid minimum bonded: less than 0: should be: 0 < inflation < 1",
			types.Params{
				MinInflation: sdk.MustNewDecFromStr("0.2"),
				MaxInflation: sdk.MustNewDecFromStr("0.2"),
				MinBonded:    sdk.MustNewDecFromStr("2"),
			},
			true,
		},
		{
			"invalid maximum bonded: less than 0: should be: 0 < inflation < 1",
			types.Params{
				MinInflation: sdk.MustNewDecFromStr("0.2"),
				MaxInflation: sdk.MustNewDecFromStr("0.2"),
				MinBonded:    sdk.MustNewDecFromStr("0.2"),
				MaxBonded:    sdk.MustNewDecFromStr("2"),
			},
			true,
		},
		{
			"invalid inflation change: less than 0: should be: 0 < inflation < 1",
			types.Params{
				MinInflation:    sdk.MustNewDecFromStr("0.2"),
				MaxInflation:    sdk.MustNewDecFromStr("0.2"),
				MinBonded:       sdk.MustNewDecFromStr("0.2"),
				MaxBonded:       sdk.MustNewDecFromStr("0.2"),
				InflationChange: sdk.MustNewDecFromStr("2"),
			},
			true,
		},
		{
			"invalid max block duration: should not be less than 0",
			types.Params{
				MinInflation:     sdk.MustNewDecFromStr("0.2"),
				MaxInflation:     sdk.MustNewDecFromStr("0.2"),
				MinBonded:        sdk.MustNewDecFromStr("0.2"),
				MaxBonded:        sdk.MustNewDecFromStr("0.2"),
				InflationChange:  sdk.MustNewDecFromStr("0.2"),
				MaxBlockDuration: -1,
			},
			true,
		},
		{
			"invalid inflation recipients: no recipients",
			types.Params{
				MinInflation:     sdk.MustNewDecFromStr("0.2"),
				MaxInflation:     sdk.MustNewDecFromStr("0.2"),
				MinBonded:        sdk.MustNewDecFromStr("0.2"),
				MaxBonded:        sdk.MustNewDecFromStr("0.2"),
				InflationChange:  sdk.MustNewDecFromStr("0.2"),
				MaxBlockDuration: time.Hour,
			},
			true,
		},
		{
			"invalid inflation recipients: ratio doesnt add up to 1",
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
						Ratio:     sdk.MustNewDecFromStr("0.2"),
					},
				},
			},
			true,
		},
		{
			"invalid inflation recipients: ratio add up to greater than 1",
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
						Ratio:     sdk.MustNewDecFromStr("0.2"),
					},
					{
						Recipient: authtypes.FeeCollectorName,
						Ratio:     sdk.MustNewDecFromStr("0.9"),
					},
				},
			},
			true,
		},
		{
			"ok: valid",
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
						Ratio:     sdk.MustNewDecFromStr("0.2"),
					},
					{
						Recipient: authtypes.FeeCollectorName,
						Ratio:     sdk.MustNewDecFromStr("0.8"),
					},
				},
			},
			false,
		},
		{
			"ok: valid: default params",
			types.DefaultParams(),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testCase, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
