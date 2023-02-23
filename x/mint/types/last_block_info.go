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
	return nil
}
