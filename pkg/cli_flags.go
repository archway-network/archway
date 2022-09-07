package pkg

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/dvsekhvalnov/jose2go/base64url"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ReadPageRequest reads and builds the necessary page request flags for pagination.
// This is fixed version of the client.ReadPageRequest function.
// The original version uses the "--page-key" flag as is, instead of base64 decoding it.
func ReadPageRequest(flagSet *pflag.FlagSet) (*query.PageRequest, error) {
	pageKeyBz, _ := flagSet.GetString(flags.FlagPageKey)
	offset, _ := flagSet.GetUint64(flags.FlagOffset)
	limit, _ := flagSet.GetUint64(flags.FlagLimit)
	countTotal, _ := flagSet.GetBool(flags.FlagCountTotal)
	page, _ := flagSet.GetUint64(flags.FlagPage)
	reverse, _ := flagSet.GetBool(flags.FlagReverse)

	if page > 1 && offset > 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "page and offset cannot be used together")
	}

	if page > 1 {
		offset = (page - 1) * limit
	}

	var pageKey []byte
	if pageKeyBz != "" {
		key, err := base64url.Decode(pageKeyBz)
		if err != nil {
			return nil, fmt.Errorf("parsing page key: %w", err)
		}
		pageKey = key
	}

	return &query.PageRequest{
		Key:        pageKey,
		Offset:     offset,
		Limit:      limit,
		CountTotal: countTotal,
		Reverse:    reverse,
	}, nil
}

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
