package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	type tc struct {
		genesis     *GenesisState
		errContains string // if empty, no error.
	}

	alice := sdk.AccAddress("alice")
	bob := sdk.AccAddress("bob")

	tests := map[string]tc{
		"ok": {
			genesis: &GenesisState{GrantingContracts: []string{alice.String(), bob.String()}},
		},
		"duplicates": {
			genesis:     &GenesisState{GrantingContracts: []string{alice.String(), alice.String()}},
			errContains: "duplicate",
		},
		"invalid addr": {
			genesis:     &GenesisState{GrantingContracts: []string{alice.String(), "invalid-address"}},
			errContains: "decoding bech32 failed",
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			err := test.genesis.Validate()
			if test.errContains == "" {
				require.Nilf(t, err, "unexpected error %s", err)
			} else {
				require.ErrorContains(t, err, test.errContains)
			}
		})
	}
}
