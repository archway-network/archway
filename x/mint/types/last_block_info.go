package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetInflationAsDec fetches the last block inflation as an sdk.Dec obj
func (lbi LastBlockInfo) GetInflationAsDec() (sdk.Dec, error) {
	inflation := lbi.GetInflation()
	inflationDec, err := sdk.NewDecFromStr(inflation)
	return inflationDec, err
}

// Validate validates all the fields in the LastBlockInfo obj
func (lbi LastBlockInfo) Validate() error {
	inflation, err := lbi.GetInflationAsDec()
	if err != nil {
		return err
	}

	if inflation.LT(sdk.ZeroDec()) || inflation.GT(sdk.OneDec()) {
		return ErrInvalidInflation
	}

	if lbi.Time == nil {
		return ErrInvalidTimestamp
	}

	return nil
}
