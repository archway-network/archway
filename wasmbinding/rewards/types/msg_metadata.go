package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

// UpdateContractMetadataRequest is the Msg.UpdateMetadata request.
type UpdateContractMetadataRequest struct {
	// ContractAddress if not empty, specifies the target contract.
	ContractAddress string `json:"contract_address"`
	// OwnerAddress if not empty, changes the contract metadata ownership.
	OwnerAddress string `json:"owner_address"`
	// RewardsAddress if not empty, changes the rewards distribution destination address.
	RewardsAddress string `json:"rewards_address"`
}

// Validate performs request fields validation.
func (r UpdateContractMetadataRequest) Validate() error {
	changeCnt := 0

	if r.ContractAddress != "" {
		if _, err := sdk.AccAddressFromBech32(r.ContractAddress); err != nil {
			return fmt.Errorf("contractAddress: parsing: %w", err)
		}
	}

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

// ToSDK convert the UpdateMetadataRequest to a rewardsTypes.Metadata.
func (r UpdateContractMetadataRequest) ToSDK() rewardsTypes.ContractMetadata {
	return rewardsTypes.ContractMetadata{
		OwnerAddress:   r.OwnerAddress,
		RewardsAddress: r.RewardsAddress,
	}
}

// MustGetContractAddressOk returns the target contract address as sdk.AccAddress if set.
// CONTRACT: panics in case of an error.
func (r UpdateContractMetadataRequest) MustGetContractAddressOk() (sdk.AccAddress, bool) {
	if r.ContractAddress == "" {
		return nil, false
	}

	addr, err := sdk.AccAddressFromBech32(r.ContractAddress)
	if err != nil {
		// Should not happen since we validate the request before this call
		panic(fmt.Errorf("wasm bindings: meta update: parsing contractAddress: %w", err))
	}

	return addr, true
}

// MustGetOwnerAddressOk returns the contract owner address as sdk.AccAddress if set to be updated.
// CONTRACT: panics in case of an error.
func (r UpdateContractMetadataRequest) MustGetOwnerAddressOk() (*sdk.AccAddress, bool) {
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
func (r UpdateContractMetadataRequest) MustGetRewardsAddressOk() (*sdk.AccAddress, bool) {
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
