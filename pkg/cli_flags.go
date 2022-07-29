package pkg

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
)

// ParseAccAddressFlag is a helper function to parse an account address CLI flag.
func ParseAccAddressFlag(cmd *cobra.Command, flagName string, isRequired bool) (*sdk.AccAddress, error) {
	v, err := cmd.Flags().GetString(flagName)
	if err != nil {
		return nil, fmt.Errorf("parsing %s flag: %w", flagName, err)
	}

	if v == "" {
		if isRequired {
			return nil, fmt.Errorf("parsing %s flag: value is required", flagName)
		}
		return nil, nil
	}

	addr, err := sdk.AccAddressFromBech32(v)
	if err != nil {
		return nil, fmt.Errorf("parsing %s flag: invalid address: %w", flagName, err)
	}

	return &addr, nil
}
