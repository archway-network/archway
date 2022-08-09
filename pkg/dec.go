package pkg

import sdk "github.com/cosmos/cosmos-sdk/types"

// NewDecFromUint64 converts a uint64 value to the sdk.Dec.
func NewDecFromUint64(v uint64) sdk.Dec {
	return sdk.NewDecFromInt(sdk.NewIntFromUint64(v))
}
