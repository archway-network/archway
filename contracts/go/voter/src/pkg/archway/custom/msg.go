package custom

import (
	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"
)

// TODO: this should be a part of Archway CW SDK. Added here as an example of how to use Custom msg.
type (
	// CustomMsg defines the Archway custom plugin message.
	CustomMsg struct {
		Rewards *RewardsMsg `json:",omitempty"`
	}

	RewardsMsg struct {
		// UpdateMetadata updates the contract rewards metadata.
		// Authorized if metadata exists for this contract and the contract address is set for the meta's DeveloperAddress field.
		UpdateMetadata *UpdateMetadataRequest

		// WithdrawRewards is a request to withdraw rewards for the contract.
		// Contract address is used as the rewards address (metadata field).
		WithdrawRewards *WithdrawRewardsRequest
	}
)

type (
	UpdateMetadataRequest struct {
		// OwnerAddress if not empty, changes the contract metadata ownership.
		OwnerAddress string
		// RewardsAddress if not empty, changes the rewards distribution destination address.
		RewardsAddress string
	}
)

type (
	WithdrawRewardsRequest struct {
		// RecordsLimit defines the maximum number of RewardsRecord objects to process.
		// Limit should not exceed the MaxWithdrawRecords param value.
		// Only one of (RecordsLimit, RecordIDs) should be set.
		RecordsLimit uint64
		// RecordIDs defines specific RewardsRecord object IDs to process.
		// Only one of (RecordsLimit, RecordIDs) should be set.
		RecordIds []uint64
	}

	WithdrawRewardsResponse struct {
		// RecordsNum is the number of RewardsRecord objects processed by the request.
		RecordsNum uint64
		// TotalRewards are the total rewards distributed.
		TotalRewards []stdTypes.Coin
	}
)
