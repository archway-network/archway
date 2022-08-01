package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"sigs.k8s.io/yaml"
)

// HasGasUsage returns true if the transaction has contract operations.
func (m TxInfo) HasGasUsage() bool {
	return m.TotalGas > 0
}

// Validate performs object fields validation.
func (m TxInfo) Validate() error {
	if m.Id == 0 {
		return fmt.Errorf("id: must be GT 0")
	}

	return nil
}

// String implements the fmt.Stringer interface.
func (m TxInfo) String() string {
	bz, _ := yaml.Marshal(m)
	return string(bz)
}

// GasUsed returns the total gas used by the operation and the flag that indicates whether operation was a noop operation.
func (m ContractOperationInfo) GasUsed() (uint64, bool) {
	gasUsed := m.VmGas + m.SdkGas
	return gasUsed, gasUsed > 0
}

// MustGetContractAddress returns the contract address.
// CONTRACT: panics on parsing error.
func (m ContractOperationInfo) MustGetContractAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.ContractAddress)
	if err != nil {
		panic(fmt.Errorf("parsing contract address (%s): %w", m.ContractAddress, err))
	}

	return addr
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

	if _, found := ContractOperation_name[int32(m.OperationType)]; !found {
		return fmt.Errorf("operationType: unknown type")
	}

	return nil
}

// String implements the fmt.Stringer interface.
func (m ContractOperationInfo) String() string {
	bz, _ := yaml.Marshal(m)
	return string(bz)
}

// String implements the fmt.Stringer interface.
func (m BlockTracking) String() string {
	bz, _ := yaml.Marshal(m)
	return string(bz)
}
