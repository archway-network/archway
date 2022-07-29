package types

import (
	"fmt"

	rewardsTypes "github.com/archway-network/archway/x/rewards/types"

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
		// OwnerAddress is the address of the contract owner (the one who can modify rewards parameters).
		OwnerAddress string `json:"owner_address"`
		// RewardsAddress is the target address for rewards distribution.
		RewardsAddress string `json:"rewards_address"`
	}
)

// NewContractMetadataResponse converts rewardsTypes.ContractMetadata to ContractMetadataResponse.
func NewContractMetadataResponse(meta rewardsTypes.ContractMetadata) ContractMetadataResponse {
	return ContractMetadataResponse{
		OwnerAddress:   meta.OwnerAddress,
		RewardsAddress: meta.RewardsAddress,
	}
}

// Validate performs request fields validation.
func (r ContractMetadataRequest) Validate() error {
	if _, err := sdk.AccAddressFromBech32(r.ContractAddress); err != nil {
		return fmt.Errorf("contractAddress: parsing: %w", err)
	}

	return nil
}

// MustGetContractAddress returns the contract address as sdk.AccAddress.
// CONTRACT: panics in case of an error.
func (r ContractMetadataRequest) MustGetContractAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(r.ContractAddress)
	if err != nil {
		// Should not happen since we validate the request before this call
		panic(fmt.Errorf("wasm bindings: meta request: parsing contractAddress: %w", err))
	}

	return addr
}
