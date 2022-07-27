package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate perform object fields validation.
func (m GenesisState) Validate() error {
	if err := m.Params.Validate(); err != nil {
		return fmt.Errorf("params: %w", err)
	}

	contractAddrSet := make(map[string]struct{})
	for i, meta := range m.ContractsMetadata {
		if err := meta.Validate(); err != nil {
			return fmt.Errorf("contractsMetadata [%d]: %w", i, err)
		}

		if _, ok := contractAddrSet[meta.ContractAddress]; ok {
			return fmt.Errorf("contractsMetadata [%d]: duplicated contract address: %s", i, meta.ContractAddress)
		}
		contractAddrSet[meta.ContractAddress] = struct{}{}
	}

	return nil
}

// Validate perform object fields validation.
func (m GenesisContractMetadata) Validate() error {
	if _, err := sdk.AccAddressFromBech32(m.ContractAddress); err != nil {
		return fmt.Errorf("contractAddress: %w", err)
	}
	if err := m.Metadata.Validate(); err != nil {
		return fmt.Errorf("metadata: %w", err)
	}

	return nil
}

// MustGetContractAddress returns the contract address parsed.
// Contract: panics of error.
func (m GenesisContractMetadata) MustGetContractAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.ContractAddress)
	if err != nil {
		panic(fmt.Errorf("invalid contract address: %w", err))
	}

	return addr
}
