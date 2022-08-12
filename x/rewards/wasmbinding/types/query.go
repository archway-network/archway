package types

import (
	"fmt"

	rewardsTypes "github.com/archway-network/archway/x/rewards/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Query is a container for custom WASM queries (one of).
type Query struct {
	// Metadata returns the contract metadata.
	Metadata *ContractMetadataRequest `json:"metadata"`

	// CurrentRewards returns the total amount of credited and ready for withdrawal rewards for an account.
	CurrentRewards *CurrentRewardsRequest `json:"current_rewards"`
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

type (
	CurrentRewardsRequest struct {
		// RewardsAddress is the bech32 encoded account address (might be the contract address as well).
		RewardsAddress string `json:"rewards_address"`
	}

	// CurrentRewardsResponse is the Query.CurrentRewards response.
	CurrentRewardsResponse struct {
		// Rewards are the total rewards eligible for withdrawal [serialized to string sdk.Coins].
		Rewards string `json:"rewards"`
	}
)

// Validate validates the query fields.
func (q Query) Validate() error {
	cnt := 0

	if q.Metadata != nil {
		if err := q.Metadata.Validate(); err != nil {
			return fmt.Errorf("metadata: %w", err)
		}
		cnt++
	}

	if q.CurrentRewards != nil {
		if err := q.CurrentRewards.Validate(); err != nil {
			return fmt.Errorf("currentRewards: %w", err)
		}
		cnt++
	}

	if cnt != 1 {
		return fmt.Errorf("one and only one field must be set")
	}

	return nil
}

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
// CONTRACT: panics in case of an error (should not happen since we validate the request).
func (r ContractMetadataRequest) MustGetContractAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(r.ContractAddress)
	if err != nil {
		// Should not happen since we validate the request before this call
		panic(fmt.Errorf("wasm bindings: meta request: parsing contractAddress: %w", err))
	}

	return addr
}

// NewCurrentRewardsResponse builds a new CurrentRewardsResponse.
func NewCurrentRewardsResponse(rewards sdk.Coins) CurrentRewardsResponse {
	return CurrentRewardsResponse{
		Rewards: rewards.String(),
	}
}

// Validate performs request fields validation.
func (r CurrentRewardsRequest) Validate() error {
	if _, err := sdk.AccAddressFromBech32(r.RewardsAddress); err != nil {
		return fmt.Errorf("rewardsAddress: parsing: %w", err)
	}

	return nil
}

// MustGetRewardsAddress returns the rewards address as sdk.AccAddress.
// CONTRACT: panics in case of an error (should not happen since we validate the request).
func (r CurrentRewardsRequest) MustGetRewardsAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(r.RewardsAddress)
	if err != nil {
		// Should not happen since we validate the request before this call
		panic(fmt.Errorf("wasm bindings: currentRewards request: parsing rewardsAddress: %w", err))
	}

	return addr
}
