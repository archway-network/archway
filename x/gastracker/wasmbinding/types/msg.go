package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Msg struct {
	// UpdateMetadata is a request to update the contract metadata.
	// Request is authorized only if the contract address is set as the DeveloperAddress (metadata field).
	UpdateMetadata *UpdateMetadataRequest `json:"update_metadata"`
}

// Validate validates the msg fields.
func (m Msg) Validate() error {
	cnt := 0

	if m.UpdateMetadata != nil {
		if err := m.UpdateMetadata.Validate(); err != nil {
			return fmt.Errorf("updateMetadata: %w", err)
		}
		cnt++
	}

	if cnt == 0 {
		return fmt.Errorf("empty msg")
	}

	return nil
}

// UpdateMetadataRequest is the Msg.SetMetadata request.
type UpdateMetadataRequest struct {
	// DeveloperAddress if not empty, changes the contract metadata ownership.
	DeveloperAddress string `json:"developer_address"`
	// RewardAddress if not empty, changes the rewards distribution destination address.
	RewardAddress string `json:"reward_address"`
}

// Validate performs request fields validation.
func (r UpdateMetadataRequest) Validate() error {
	changeCnt := 0

	if r.DeveloperAddress != "" {
		if _, err := sdk.AccAddressFromBech32(r.DeveloperAddress); err != nil {
			return fmt.Errorf("developerAddress: parsing: %w", err)
		}
		changeCnt++
	}

	if r.RewardAddress != "" {
		if _, err := sdk.AccAddressFromBech32(r.RewardAddress); err != nil {
			return fmt.Errorf("rewardAddress: parsing: %w", err)
		}
		changeCnt++
	}

	if changeCnt == 0 {
		return fmt.Errorf("empty request")
	}

	return nil
}

// GetDeveloperAddress returns the contract developer address as sdk.AccAddress if set to be updated.
func (r UpdateMetadataRequest) GetDeveloperAddress() (string, bool) {
	if r.DeveloperAddress == "" {
		return "", false
	}

	return r.DeveloperAddress, true
}

// GetRewardAddress returns the contract rewards address as sdk.AccAddress if set to be updated.
func (r UpdateMetadataRequest) GetRewardAddress() (string, bool) {
	if r.RewardAddress == "" {
		return "", false
	}

	return r.RewardAddress, true
}
