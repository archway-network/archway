package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramTypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramTypes.ParamSet = (*Params)(nil)

// ParamKeyTable creates a new params table.
func ParamKeyTable() paramTypes.KeyTable {
	return paramTypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance.
func NewParams(inflationRewardsRatio, txFeeRebateRatio sdk.Dec, maxwithdrawRecords uint64) Params {
	panic("unimplemented 👻")
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	panic("unimplemented 👻")
}

// ParamSetPairs Implements the paramTypes.ParamSet interface.
func (m *Params) ParamSetPairs() paramTypes.ParamSetPairs {
	panic("unimplemented 👻")
}

// Validate perform object fields validation.
func (m Params) Validate() error {
	panic("unimplemented 👻")
}
