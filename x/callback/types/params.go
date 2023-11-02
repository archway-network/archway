package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewParams creates a new Params instance.
func NewParams(inflationRewardsRatio, txFeeRebateRatio sdk.Dec, maxwithdrawRecords uint64) Params {
	panic("unimplemented 👻")
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	panic("unimplemented 👻")
}

// Validate perform object fields validation.
func (m Params) Validate() error {
	panic("unimplemented 👻")
}
