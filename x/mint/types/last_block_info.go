package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (lbi LastBlockInfo) GetInflationAsDec() (sdk.Dec, error) {
	inflation := lbi.GetInflation()
	inflationDec, err := sdk.NewDecFromStr(inflation)
	return inflationDec, err
}

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
