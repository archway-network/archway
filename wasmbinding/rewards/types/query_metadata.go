package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

// ContractMetadataRequest is the Query.Metadata request.
type ContractMetadataRequest struct {
	// ContractAddress is the bech32 encoded contract address.
	ContractAddress string `json:"contract_address"`
}

// ContractMetadataResponse is the Query.Metadata response.
type ContractMetadataResponse struct {
	// OwnerAddress is the address of the contract owner (the one who can modify rewards parameters).
	OwnerAddress string `json:"owner_address"`
	// RewardsAddress is the target address for rewards distribution.
	RewardsAddress string `json:"rewards_address"`
}

// Validate performs request fields validation.
func (r ContractMetadataRequest) Validate() error {
	if _, err := sdk.AccAddressFromBech32(r.ContractAddress); err != nil {
		return fmt.Errorf("contractAddress: parsing: %w", err)
	}

	return nil
}

// MustGetContractAddress returns the contract address as sdk.AccAddress.
// CONTRACT: panics in case of an error (should not happen since we validate the request).
func (r ContractMetadataRequest) MustGetContractAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(r.ContractAddress)
	if err != nil {
		// Should not happen since we validate the request before this call
		panic(fmt.Errorf("wasm bindings: meta request: parsing contractAddress: %w", err))
	}

	return addr
}

// NewContractMetadataResponse converts rewardsTypes.ContractMetadata to ContractMetadataResponse.
func NewContractMetadataResponse(meta rewardsTypes.ContractMetadata) ContractMetadataResponse {
	return ContractMetadataResponse{
		OwnerAddress:   meta.OwnerAddress,
		RewardsAddress: meta.RewardsAddress,
	}
}
