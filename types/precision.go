package types

import (
	"math/big"

	math "cosmossdk.io/math"
)

const (
	BaseDenomUnit = 18
)

var DefaultPowerReduction = math.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(BaseDenomUnit), nil)) // 10^18
