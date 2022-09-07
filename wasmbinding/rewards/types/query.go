package types

import (
	"fmt"
)

// Query is a container for custom WASM query for the x/rewards module (one of).
type Query struct {
	// Metadata returns the contract metadata.
	Metadata *ContractMetadataRequest `json:"metadata"`

	// RewardsRecords returns a list of RewardsRecord objects that are credited for the account and are ready to be withdrawn.
	// Request is paginated. If the limit field is not set, the MaxWithdrawRecords param is used.
	RewardsRecords *RewardsRecordsRequest `json:"rewards_records"`
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

	if q.RewardsRecords != nil {
		if err := q.RewardsRecords.Validate(); err != nil {
			return fmt.Errorf("rewardsRecords: %w", err)
		}
		cnt++
	}

	if cnt != 1 {
		return fmt.Errorf("one and only one field must be set")
	}

	return nil
}
