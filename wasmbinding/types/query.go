package types

import (
	"fmt"

	govTypes "github.com/archway-network/archway/wasmbinding/gov/types"
	rewardsTypes "github.com/archway-network/archway/wasmbinding/rewards/types"
)

// Query is a container for custom WASM queries (one of).
type Query struct {
	// ContractMetadata returns the contract metadata.
	ContractMetadata *rewardsTypes.ContractMetadataRequest `json:"contract_metadata"`

	// RewardsRecords returns a list of RewardsRecord objects that are credited for the account and are ready to be withdrawn.
	// The request is paginated. If the limit field is not set, the MaxWithdrawRecords param is used.
	RewardsRecords *rewardsTypes.RewardsRecordsRequest `json:"rewards_records"`

	// GovVote returns the vote data for a given proposal and voter.
	GovVote *govTypes.VoteRequest `json:"gov_vote"`

	// FlatFee returns the contracts flat fee
	FlatFee *rewardsTypes.ContractFlatFeeRequest `json:"flat_fee"`
}

// Validate validates the query fields.
func (q Query) Validate() error {
	cnt := 0

	if q.ContractMetadata != nil {
		cnt++
	}

	if q.RewardsRecords != nil {
		cnt++
	}

	if q.GovVote != nil {
		cnt++
	}

	if q.FlatFee != nil {
		cnt++
	}

	if cnt != 1 {
		return fmt.Errorf("one and only one sub-query must be set (fields=%v)", cnt)
	}

	return nil
}
