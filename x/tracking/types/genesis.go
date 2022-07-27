package types

import "fmt"

// NewGenesisState creates a new GenesisState object.
func NewGenesisState(params Params, txInfos []TxInfo, contractOpInfos []ContractOperationInfo) *GenesisState {
	return &GenesisState{
		Params:          params,
		TxInfos:         txInfos,
		ContractOpInfos: contractOpInfos,
	}
}

// DefaultGenesisState returns a default genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:          DefaultParams(),
		TxInfos:         []TxInfo{},
		ContractOpInfos: []ContractOperationInfo{},
	}
}

// Validate performs genesis state validation.
func (m GenesisState) Validate() error {
	if err := m.Params.Validate(); err != nil {
		return fmt.Errorf("params: %w", err)
	}

	txIDSet := make(map[uint64]struct{})
	for i, txInfo := range m.TxInfos {
		if err := txInfo.Validate(); err != nil {
			return fmt.Errorf("txInfos [%d]: %w", i, err)
		}
		if _, ok := txIDSet[txInfo.Id]; ok {
			return fmt.Errorf("txInfos [%d]: duplicated ID: %d", i, txInfo.Id)
		}
		txIDSet[txInfo.Id] = struct{}{}
	}

	opIDSet := make(map[uint64]struct{})
	for i, opInfo := range m.ContractOpInfos {
		if err := opInfo.Validate(); err != nil {
			return fmt.Errorf("contractOpInfos [%d]: %w", i, err)
		}
		if _, ok := opIDSet[opInfo.Id]; ok {
			return fmt.Errorf("contractOpInfos [%d]: duplicated ID: %d", i, opInfo.Id)
		}
		if _, ok := txIDSet[opInfo.TxId]; !ok {
			return fmt.Errorf("contractOpInfos [%d]: txId (%d): not found", i, opInfo.TxId)
		}
		opIDSet[opInfo.Id] = struct{}{}
	}

	return nil
}
