package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"sigs.k8s.io/yaml"
)

// String implements the fmt.Stringer interface.
func (m TxInfo) String() string {
	bz, _ := yaml.Marshal(m)
	return string(bz)
}

// Validate performs object fields validation.
func (m TxInfo) Validate() error {
	if m.Id == 0 {
		return fmt.Errorf("id: must be GT 0")
	}

	return nil
}

// String implements the fmt.Stringer interface.
func (m ContractOperationInfo) String() string {
	bz, _ := yaml.Marshal(m)
	return string(bz)
}

// Validate performs object fields validation.
func (m ContractOperationInfo) Validate() error {
	if m.Id == 0 {
		return fmt.Errorf("id: must be GT 0")
	}
	if m.TxId == 0 {
		return fmt.Errorf("txId: must be GT 0")
	}
	if _, err := sdk.AccAddressFromBech32(m.ContractAddress); err != nil {
		return fmt.Errorf("contractAddress: %s", err.Error())
	}

	return nil
}
