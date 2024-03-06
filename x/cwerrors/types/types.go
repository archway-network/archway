package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate perform object fields validation.
func (s SudoError) Validate() error {
	_, err := sdk.AccAddressFromBech32(s.ContractAddress)
	if err != nil {
		return err
	}
	if s.ModuleName == "" {
		return ErrModuleNameMissing
	}
	return nil
}
