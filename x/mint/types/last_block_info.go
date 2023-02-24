package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate validates all the fields in the LastBlockInfo obj
func (lbi LastBlockInfo) Validate() error {
	if lbi.Inflation.LT(sdk.ZeroDec()) || lbi.Inflation.GT(sdk.OneDec()) {
		return ErrInvalidInflation
	}

	if lbi.Time == nil {
		return ErrInvalidTimestamp
	}

	return nil
}
