package pkg

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ParseAccAddressArg is a helper function to parse an account address CLI argument.
func ParseAccAddressArg(argName, argValue string) (sdk.AccAddress, error) {
	addr, err := sdk.AccAddressFromBech32(argValue)
	if err != nil {
		return sdk.AccAddress{}, fmt.Errorf("parsing %s argument: invalid address: %w", argName, err)
	}

	return addr, nil
}
