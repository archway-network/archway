package custom

import (
	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"
)

// This should be a part of Archway CW SDK. Added here as an example of how to use Custom queries.
type (
	// CustomQuery defines the Archway custom plugin query.
	CustomQuery struct {
		// ContractMetadata returns the contract rewards metadata.
		ContractMetadata *ContractMetadataRequest `json:",omitempty"`

		// RewardsRecords returns a list of RewardsRecord objects that are credited for the account and are ready to be withdrawn.
		// Request is paginated. If the limit field is not set, the MaxWithdrawRecords param is used.
		RewardsRecords *RewardsRecordsRequest `json:",omitempty"`

		GovVote *GovVoteRequest `json:",omitempty"`
	}
)

type (
	ContractMetadataRequest struct {
		// ContractAddress is a contract address to get metadata for.
		ContractAddress string
	}

	ContractMetadataResponse struct {
		// OwnerAddress is the address of the contract owner (the one who can modify rewards parameters).
		OwnerAddress string
		// RewardsAddress is the target address for rewards distribution.
		RewardsAddress string
	}
)

type (
	RewardsRecordsRequest struct {
		// RewardsAddress is the bech32 encoded account address (might be the contract address as well).
		RewardsAddress string
		// Pagination is an optional pagination options for the request.
		Pagination *PageRequest
	}

	RewardsRecordsResponse struct {
		// Records is the list of rewards records returned by the query.
		Records []RewardsRecord
		// Pagination is the pagination details in the response.
		Pagination PageResponse
	}

	RewardsRecord struct {
		// ID is the unique ID of the record.
		ID uint64
		// RewardsAddress is the address to distribute rewards to (bech32 encoded).
		RewardsAddress string
		// Rewards are the rewards to be transferred later.
		Rewards []stdTypes.Coin
		// CalculatedHeight defines the block height of rewards calculation event.
		CalculatedHeight int64
		// CalculatedTime defines the block time of rewards calculation event.
		// RFC3339Nano is used to represent the time.
		CalculatedTime string
	}
)

type ( // NOTE: only GovVoteRequest is present, for testing we don't care about the response, this is to be maintained in a shared repo and not be replicated.
	GovVoteRequest struct {
		ProposalID uint64 `json:",omitempty"`
		Voter      string `json:",omitempty"`
	}
)

type (
	PageRequest struct {
		// Key is a value returned in the PageResponse.NextKey to begin querying the next page most efficiently.
		// Only one of (Offset, Key) should be set.
		// Go converts a byte slice into a base64 encoded string, so this field could be a string in a contract code.
		Key []byte
		// Offset is a numeric offset that can be used when Key is unavailable.
		// Only one of (Offset, Key) should be set.
		Offset uint64
		// Limit is the total number of results to be returned in the result page.
		// It is set to default if left empty.
		Limit uint64
		// CountTotal if set to true, indicates that the result set should include a count of the total number of items available for pagination.
		// This field is only respected when the Offset is used.
		// It is ignored when Key field is set.
		CountTotal bool
		// Reverse if set to true, results are to be returned in the descending order.
		Reverse bool
	}

	PageResponse struct {
		// NextKey is the key to be passed to PageRequest.Key to query the next page most efficiently.
		NextKey []byte
		// Total is the total number of results available if PageRequest.CountTotal was set, its value is undefined otherwise.
		Total uint64
	}
)
