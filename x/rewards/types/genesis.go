package types

import (
	"fmt"
)

// NewGenesisState creates a new GenesisState object.
func NewGenesisState(params Params, contractsMetadata []ContractMetadata, blockRewards []BlockRewards, txRewards []TxRewards) *GenesisState {
	return &GenesisState{
		Params:            params,
		ContractsMetadata: contractsMetadata,
		BlockRewards:      blockRewards,
		TxRewards:         txRewards,
	}
}

// DefaultGenesisState returns a default genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:            DefaultParams(),
		ContractsMetadata: []ContractMetadata{},
		BlockRewards:      []BlockRewards{},
		TxRewards:         []TxRewards{},
	}
}

// Validate perform object fields validation.
func (m GenesisState) Validate() error {
	if err := m.Params.Validate(); err != nil {
		return fmt.Errorf("params: %w", err)
	}

	contractAddrSet := make(map[string]struct{})
	for i, meta := range m.ContractsMetadata {
		if err := meta.Validate(true); err != nil {
			return fmt.Errorf("contractsMetadata [%d]: %w", i, err)
		}

		if _, ok := contractAddrSet[meta.ContractAddress]; ok {
			return fmt.Errorf("contractsMetadata [%d]: duplicated contract address: %s", i, meta.ContractAddress)
		}
		contractAddrSet[meta.ContractAddress] = struct{}{}
	}

	blockRewardsHeightSet := make(map[int64]struct{})
	for i, blockRewards := range m.BlockRewards {
		if err := blockRewards.Validate(); err != nil {
			return fmt.Errorf("blockRewards [%d]: %w", i, err)
		}
		if _, ok := blockRewardsHeightSet[blockRewards.Height]; ok {
			return fmt.Errorf("blockRewards [%d]: duplicated height: %d", i, blockRewards.Height)
		}
		blockRewardsHeightSet[blockRewards.Height] = struct{}{}
	}

	txRewardsIdSet := make(map[uint64]struct{})
	for i, txRewards := range m.TxRewards {
		if err := txRewards.Validate(); err != nil {
			return fmt.Errorf("txRewards [%d]: %w", i, err)
		}
		if _, ok := blockRewardsHeightSet[txRewards.Height]; !ok {
			return fmt.Errorf("txRewards [%d]: height not found: %d", i, txRewards.Height)
		}
		if _, ok := txRewardsIdSet[txRewards.TxId]; ok {
			return fmt.Errorf("txRewards [%d]: duplicated txId: %d", i, txRewards.TxId)
		}
		txRewardsIdSet[txRewards.TxId] = struct{}{}
	}

	return nil
}
