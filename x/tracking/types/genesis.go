package types

import "fmt"

// NewGenesisState creates a new GenesisState object.
func NewGenesisState(txInfoLastID uint64, txInfos []TxInfo, contractOpInfoLastID uint64, contractOpInfos []ContractOperationInfo) *GenesisState {
	return &GenesisState{
		TxInfoLastId:         txInfoLastID,
		TxInfos:              txInfos,
		ContractOpInfoLastId: contractOpInfoLastID,
		ContractOpInfos:      contractOpInfos,
	}
}

// DefaultGenesisState returns a default genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		TxInfoLastId:         0,
		TxInfos:              []TxInfo{},
		ContractOpInfoLastId: 0,
		ContractOpInfos:      []ContractOperationInfo{},
	}
}

// Validate performs genesis state validation.
func (m GenesisState) Validate() error {
	txIDMax := uint64(0)
	txIDSet := make(map[uint64]struct{})
	for i, txInfo := range m.TxInfos {
		if err := txInfo.Validate(); err != nil {
			return fmt.Errorf("txInfos [%d]: %w", i, err)
		}
		if _, ok := txIDSet[txInfo.Id]; ok {
			return fmt.Errorf("txInfos [%d]: duplicated ID: %d", i, txInfo.Id)
		}

		if txInfo.Id > txIDMax {
			txIDMax = txInfo.Id
		}
		txIDSet[txInfo.Id] = struct{}{}
	}

	if m.TxInfoLastId < txIDMax {
		return fmt.Errorf("txInfoLastId: %d < max TxInfo ID (%d)", m.TxInfoLastId, txIDMax)
	}

	opIDMax := uint64(0)
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

		if opInfo.Id > opIDMax {
			opIDMax = opInfo.Id
		}
		opIDSet[opInfo.Id] = struct{}{}
	}

	if m.ContractOpInfoLastId < opIDMax {
		return fmt.Errorf("contractOpInfoLastId: %d < max ContractOpInfo ID (%d)", m.ContractOpInfoLastId, opIDMax)
	}

	return nil
}
