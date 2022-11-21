package types

import (
	"fmt"

	rewardsTypes "github.com/archway-network/archway/wasmbinding/rewards/types"
)

// Msg is a container for custom WASM message operations in all of Archway's custom modules.
type Msg struct {
	// UpdateContractMetadata is a request to update the contract metadata.
	// Request is authorized only if the contract address is set as the DeveloperAddress (metadata field).
	UpdateContractMetadata *rewardsTypes.UpdateContractMetadataRequest `json:"update_rewards_metadata"`

	// WithdrawRewards is a request to withdraw rewards for the contract.
	// Contract address is used as the rewards address (metadata field).
	WithdrawRewards *rewardsTypes.WithdrawRewardsRequest `json:"withdraw_rewards"`
}

// Validate validates the msg fields.
func (m Msg) Validate() error {
	cnt := 0

	if m.UpdateContractMetadata != nil {
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
