package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math/big"
)

const (
	BaseDenomUnit = 18
)

var (
	DefaultPowerReduction = sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(BaseDenomUnit), nil)) // 10^18
)
