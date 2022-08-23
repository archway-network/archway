package types

import (
	"fmt"
)

// Msg is a container for custom WASM message operations for the x/rewards module (one of).
type Msg struct {
	// UpdateMetadata is a request to update the contract metadata.
	// Request is authorized only if the contract address is set as the DeveloperAddress (metadata field).
	UpdateMetadata *UpdateMetadataRequest `json:"update_metadata"`

	// WithdrawRewards is a request to withdraw rewards for the contract.
	// Contract address is used as the rewards address (metadata field).
	WithdrawRewards *WithdrawRewardsRequest `json:"withdraw_rewards"`
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

	if m.WithdrawRewards != nil {
		if err := m.WithdrawRewards.Validate(); err != nil {
			return fmt.Errorf("withdrawRewards: %w", err)
		}
		cnt++
	}

	if cnt != 1 {
		return fmt.Errorf("one and only one field must be set")
	}

	return nil
}
