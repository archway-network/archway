package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/pkg"
)

// NewGenesisState creates a new GenesisState object.
func NewGenesisState(
	params Params,
	contractsMetadata []ContractMetadata,
	blockRewards []BlockRewards, txRewards []TxRewards,
	minConsFee sdk.DecCoin,
	rewardsRecordLastID uint64,
	rewardsRecords []RewardsRecord,
	flatFees []FlatFee,
) *GenesisState {
	return &GenesisState{
		Params:              params,
		ContractsMetadata:   contractsMetadata,
		BlockRewards:        blockRewards,
		TxRewards:           txRewards,
		MinConsensusFee:     minConsFee,
		RewardsRecordLastId: rewardsRecordLastID,
		RewardsRecords:      rewardsRecords,
		FlatFees:            flatFees,
	}
}

// DefaultGenesisState returns a default genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:              DefaultParams(),
		ContractsMetadata:   []ContractMetadata{},
		BlockRewards:        []BlockRewards{},
		TxRewards:           []TxRewards{},
		MinConsensusFee:     sdk.DecCoin{},
		RewardsRecordLastId: 0,
		RewardsRecords:      []RewardsRecord{},
		FlatFees:            []FlatFee{},
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

	if !pkg.DecCoinIsZero(m.MinConsensusFee) {
		if err := pkg.ValidateDecCoin(m.MinConsensusFee); err != nil {
			return fmt.Errorf("minConsensusFee: %w", err)
		}
	}

	rewardsRecordIDMax := uint64(0)
	rewardsRecordsIdSet := make(map[uint64]struct{})
	for i, rewardsRecord := range m.RewardsRecords {
		if err := rewardsRecord.Validate(); err != nil {
			return fmt.Errorf("rewardsRecords [%d]: %w", i, err)
		}
		if _, ok := rewardsRecordsIdSet[rewardsRecord.Id]; ok {
			return fmt.Errorf("rewardsRecords [%d]: duplicated id: %d", i, rewardsRecord.Id)
		}

		if rewardsRecord.Id > rewardsRecordIDMax {
			rewardsRecordIDMax = rewardsRecord.Id
		}
		rewardsRecordsIdSet[rewardsRecord.Id] = struct{}{}
	}

	if m.RewardsRecordLastId < rewardsRecordIDMax {
		return fmt.Errorf("rewardsRecordLastId: %d < max RewardsRecord ID (%d)", m.RewardsRecordLastId, rewardsRecordIDMax)
	}

	flatFeeSet := make(map[string]struct{})
	for i, fee := range m.FlatFees {
		if _, ok := contractAddrSet[fee.ContractAddress]; !ok {
			return fmt.Errorf("flat fee: %+v is invalid, err: contract metadata not found", fee)
		}

		if err := fee.Validate(); err != nil {
			return fmt.Errorf("flatFee [%d]: %w", i, err)
		}

		if _, ok := flatFeeSet[fee.ContractAddress]; ok {
			return fmt.Errorf("flatFee [%d]: duplicated contract address: %s", i, fee.ContractAddress)
		}
		flatFeeSet[fee.ContractAddress] = struct{}{}
	}

	return nil
}
