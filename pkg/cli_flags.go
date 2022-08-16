package pkg

import (
	"fmt"
	"strconv"

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

// GetUint64Flag is a helper function to get a uint64 CLI flag.
func GetUint64Flag(cmd *cobra.Command, flagName string, canBeEmpty bool) (uint64, error) {
	v, err := cmd.Flags().GetUint64(flagName)
	if err != nil {
		return 0, fmt.Errorf("parsing %s flag: %w", flagName, err)
	}

	if v == 0 && !canBeEmpty {
		return 0, fmt.Errorf("parsing %s flag: value is required", flagName)
	}

	return v, nil
}

// GetStringSliceFlag is a helper function to get a slice of strings CLI flag.
func GetStringSliceFlag(cmd *cobra.Command, flagName string, canBeEmpty bool) ([]string, error) {
	v, err := cmd.Flags().GetStringSlice(flagName)
	if err != nil {
		return nil, fmt.Errorf("parsing %s flag: %w", flagName, err)
	}

	if len(v) == 0 && !canBeEmpty {
		return nil, fmt.Errorf("parsing %s flag: can not be empty", flagName)
	}

	return v, nil
}

// GetUint64SliceFlag is a helper function to get a slice of uint64 CLI flag.
func GetUint64SliceFlag(cmd *cobra.Command, flagName string, canBeEmpty bool) ([]uint64, error) {
	v, err := cmd.Flags().GetStringSlice(flagName)
	if err != nil {
		return nil, fmt.Errorf("parsing %s flag: %w", flagName, err)
	}

	if len(v) == 0 && !canBeEmpty {
		return nil, fmt.Errorf("parsing %s flag: can not be empty", flagName)
	}

	var result []uint64
	for _, valueBz := range v {
		value, err := strconv.ParseUint(valueBz, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing %s flag: invalid value (%s): %w", flagName, valueBz, err)
		}
		result = append(result, value)
	}

	return result, nil
}
