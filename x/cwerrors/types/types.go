package types

import (
	"encoding/json"

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

func (s SudoError) Bytes() []byte {
	msgBz, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return msgBz
}

// func (s SudoError) String() string {
// 	return string(s.Bytes())
// }
