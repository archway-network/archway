package pkg

import (
	math "cosmossdk.io/math"
)

// NewDecFromUint64 converts a uint64 value to the sdk.Dec.
func NewDecFromUint64(v uint64) math.LegacyDec {
	return math.LegacyNewDecFromInt(math.NewIntFromUint64(v))
}
