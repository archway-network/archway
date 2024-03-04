package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/archway-network/archway/x/cwica/types"
)

// TestGenesisState_Validate tests the Validate method of GenesisState
func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "MsgSendTxMaxMessages must be greater than zero",
			genState: &types.GenesisState{
				Params: types.Params{
					MsgSendTxMaxMessages: 0,
				},
			},
			valid: false,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				Params: types.Params{
					MsgSendTxMaxMessages: 10,
				},
			},
			valid: true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
