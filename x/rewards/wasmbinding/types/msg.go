package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

// Msg is a container for custom WASM messages (one of).
type Msg struct {
	// UpdateMetadata is a request to update the contract metadata.
	// Request is authorized only if the contract address is set as the DeveloperAddress (metadata field).
	UpdateMetadata *UpdateMetadataRequest `json:"update_metadata"`

	// WithdrawRewards is a request to withdraw rewards for the contract.
	// Contract address is used as the rewards address (metadata field).
	WithdrawRewards *WithdrawRewardsRequest `json:"withdraw_rewards"`
}

type (
	// UpdateMetadataRequest is the Msg.SetMetadata request.
	UpdateMetadataRequest struct {
		// OwnerAddress if not empty, changes the contract metadata ownership.
		OwnerAddress string `json:"owner_address"`
		// RewardsAddress if not empty, changes the rewards distribution destination address.
		RewardsAddress string `json:"rewards_address"`
	}
)

type (
	// WithdrawRewardsRequest is the Msg.WithdrawRewards request.
	WithdrawRewardsRequest struct{}

	// WithdrawRewardsResponse is the Msg.WithdrawRewards response.
	WithdrawRewardsResponse struct {
		// Rewards are the total rewards distributed [serialized to string sdk.Coins].
		Rewards string `json:"rewards"`
	}
)

// Validate validates the msg fields.
func (m Msg) Validate() error {
	cnt := 0

	if m.UpdateMetadata != nil {
		if err := m.UpdateMetadata.Validate(); err != nil {
			return fmt.Errorf("updateMetadata: %w", err)
		}
		cnt++
	}

	if m.WithdrawRewards != nil {
		cnt++
	}

	if cnt != 1 {
		return fmt.Errorf("one and only one field must be set")
	}

	return nil
}

// Validate performs request fields validation.
func (r UpdateMetadataRequest) Validate() error {
	changeCnt := 0

	if r.OwnerAddress != "" {
		if _, err := sdk.AccAddressFromBech32(r.OwnerAddress); err != nil {
			return fmt.Errorf("ownerAddress: parsing: %w", err)
		}
		changeCnt++
	}

	if r.RewardsAddress != "" {
		if _, err := sdk.AccAddressFromBech32(r.RewardsAddress); err != nil {
			return fmt.Errorf("rewardsAddress: parsing: %w", err)
		}
		changeCnt++
	}

	if changeCnt == 0 {
		return fmt.Errorf("empty request")
	}

	return nil
}

// ToMetadata convert the UpdateMetadataRequest to a rewardsTypes.Metadata.
func (r UpdateMetadataRequest) ToMetadata() rewardsTypes.ContractMetadata {
	return rewardsTypes.ContractMetadata{
		OwnerAddress:   r.OwnerAddress,
		RewardsAddress: r.RewardsAddress,
	}
}

// MustGetOwnerAddressOk returns the contract owner address as sdk.AccAddress if set to be updated.
// CONTRACT: panics in case of an error.
func (r UpdateMetadataRequest) MustGetOwnerAddressOk() (*sdk.AccAddress, bool) {
	if r.OwnerAddress == "" {
		return nil, false
	}

	addr, err := sdk.AccAddressFromBech32(r.OwnerAddress)
	if err != nil {
		// Should not happen since we validate the request before this call
		panic(fmt.Errorf("wasm bindings: meta update: parsing ownerAddress: %w", err))
	}

	return &addr, true
}

// MustGetRewardsAddressOk returns the rewards address as sdk.AccAddress if set to be updated.
func (r UpdateMetadataRequest) MustGetRewardsAddressOk() (*sdk.AccAddress, bool) {
	if r.RewardsAddress == "" {
		return nil, false
	}

	addr, err := sdk.AccAddressFromBech32(r.RewardsAddress)
	if err != nil {
		// Should not happen since we validate the request before this call
		panic(fmt.Errorf("wasm bindings: meta update: parsing rewardsAddress: %w", err))
	}

	return &addr, true
}

// NewWithdrawRewardsResponse creates a new WithdrawRewardsResponse.
func NewWithdrawRewardsResponse(rewards sdk.Coins) WithdrawRewardsResponse {
	return WithdrawRewardsResponse{
		Rewards: rewards.String(),
	}
}
