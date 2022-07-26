package types

import (
	"fmt"

	"github.com/archway-network/archway/x/gastracker"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Query is a container for custom WASM queries.
type Query struct {
	// Metadata returns the contract metadata.
	Metadata *ContractMetadataRequest `json:"metadata"`
}

// Validate validates the query fields.
func (q Query) Validate() error {
	cnt := 0

	if q.Metadata != nil {
		if err := q.Metadata.Validate(); err != nil {
			return fmt.Errorf("metadata: %w", err)
		}
		cnt++
	}

	if cnt != 1 {
		return fmt.Errorf("one and only one field must be set")
	}

	return nil
}

type (
	// ContractMetadataRequest is the Query.Metadata request.
	ContractMetadataRequest struct {
		// ContractAddress is the bech32 encoded contract address.
		ContractAddress string `json:"contract_address"`
	}

	// ContractMetadataResponse is the Query.Metadata response.
	ContractMetadataResponse struct {
		// DeveloperAddress is an address of the contract developer.
		DeveloperAddress string `json:"developer_address"`
		// RewardAddress is an address for rewards distribution.
		RewardAddress string `json:"reward_address"`
		// GasRebateToUserEnabled flag indicates whether 50% discount is enabled for using this contract (lowers gas usage).
		// One of: GasRebateToUserEnabled || PremiumEnabled
		GasRebateToUserEnabled bool `json:"gas_rebate_to_user_enabled"`
		// PremiumEnabled flag indicates whether extra charge is enabled for using this contract (raises gas usage).
		// One of: GasRebateToUserEnabled || PremiumEnabled
		PremiumEnabled bool `json:"premium_enabled"`
		// PremiumPercentage defines the premium percentage of gas consumed to be charged [0, 100].
		PremiumPercentage uint16 `json:"premium_percentage"`
	}
)

// NewContractMetadataResponse converts ContractInstanceMetadata to ContractMetadataResponse.
func NewContractMetadataResponse(meta gastracker.ContractInstanceMetadata) ContractMetadataResponse {
	return ContractMetadataResponse{
		DeveloperAddress:       meta.DeveloperAddress,
		RewardAddress:          meta.RewardAddress,
		GasRebateToUserEnabled: meta.GasRebateToUser,
		PremiumEnabled:         meta.CollectPremium,
		PremiumPercentage:      uint16(meta.PremiumPercentageCharged),
	}
}

// Validate performs request fields validation.
func (r ContractMetadataRequest) Validate() error {
	if _, err := sdk.AccAddressFromBech32(r.ContractAddress); err != nil {
		return fmt.Errorf("contractAddress: parsing: %w", err)
	}

	return nil
}

// GetContractAddress returns the contract address as sdk.AccAddress.
func (r ContractMetadataRequest) GetContractAddress() sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(r.ContractAddress)
	return addr
}
