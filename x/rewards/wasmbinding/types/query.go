package types

import (
	"fmt"
	"time"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmVmTypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/archway-network/archway/pkg"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	// Query is a container for custom WASM query.
	Query struct {
		// Rewards defines the x/rewards module specific query.
		Rewards *RewardsQuery `json:"rewards,omitempty"`
	}

	// RewardsQuery is a container for custom WASM query for the x/rewards module (one of).
	RewardsQuery struct {
		// Metadata returns the contract metadata.
		Metadata *ContractMetadataRequest `json:"metadata"`

		// RewardsRecords returns a list of RewardsRecord objects that are credited for the account and are ready to be withdrawn.
		// Request is paginated. If the limit field is not set, the MaxWithdrawRecords param is used.
		RewardsRecords *RewardsRecordsRequest `json:"rewards_records"`
	}
)

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
	// RewardsRecordsRequest is the Query.RewardsRecords request.
	RewardsRecordsRequest struct {
		// RewardsAddress is the bech32 encoded account address (might be the contract address as well).
		RewardsAddress string `json:"rewards_address"`
		// Pagination is an optional pagination options for the request.
		// Limit should not exceed the MaxWithdrawRecords param value.
		Pagination *PageRequest `json:"pagination"`
	}

	// RewardsRecordsResponse is the Query.RewardsRecords response.
	RewardsRecordsResponse struct {
		// Records is the list of rewards records returned by the query.
		Records []RewardsRecord `json:"records"`
		// Pagination is the pagination details in the response.
		Pagination PageResponse `json:"pagination"`
	}

	// RewardsRecord is the WASM binding representation of a rewardsTypes.RewardsRecord object.
	RewardsRecord struct {
		// ID is the unique ID of the record.
		ID uint64 `json:"id"`
		// RewardsAddress is the address to distribute rewards to (bech32 encoded).
		RewardsAddress string `json:"rewards_address"`
		// Rewards are the rewards to be transferred later.
		Rewards wasmVmTypes.Coins `json:"rewards"`
		// CalculatedHeight defines the block height of rewards calculation event.
		CalculatedHeight int64 `json:"calculated_height"`
		// CalculatedTime defines the block time of rewards calculation event.
		// RFC3339Nano is used to represent the time.
		CalculatedTime string `json:"calculated_time"`
	}
)

// Validate validates the query fields.
func (q Query) Validate() error {
	cnt := 0

	if q.Rewards == nil {
		return nil
	}

	if q.Rewards.Metadata != nil {
		if err := q.Rewards.Metadata.Validate(); err != nil {
			return fmt.Errorf("rewards: metadata: %w", err)
		}
		cnt++
	}

	if q.Rewards.RewardsRecords != nil {
		if err := q.Rewards.RewardsRecords.Validate(); err != nil {
			return fmt.Errorf("rewards: rewardsRecords: %w", err)
		}
		cnt++
	}

	if cnt != 1 {
		return fmt.Errorf("rewards: one and only one field must be set")
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

// ToSDK converts the RewardsRecord to rewardsTypes.RewardsRecord.
func (r RewardsRecord) ToSDK() (rewardsTypes.RewardsRecord, error) {
	rewards, err := pkg.WasmCoinsToSDK(r.Rewards)
	if err != nil {
		return rewardsTypes.RewardsRecord{}, fmt.Errorf("rewards: %w", err)
	}

	calculatedTime, err := time.Parse(time.RFC3339Nano, r.CalculatedTime)
	if err != nil {
		return rewardsTypes.RewardsRecord{}, fmt.Errorf("calculatedTime: %w", err)
	}

	return rewardsTypes.RewardsRecord{
		Id:               r.ID,
		RewardsAddress:   r.RewardsAddress,
		Rewards:          rewards,
		CalculatedHeight: r.CalculatedHeight,
		CalculatedTime:   calculatedTime,
	}, nil
}

// NewRewardsRecordsResponse builds a new RewardsRecordsResponse.
func NewRewardsRecordsResponse(records []rewardsTypes.RewardsRecord, pageResp query.PageResponse) RewardsRecordsResponse {
	resp := RewardsRecordsResponse{
		Records:    make([]RewardsRecord, 0, len(records)),
		Pagination: NewPageResponseFromSDK(pageResp),
	}

	for _, record := range records {
		resp.Records = append(resp.Records, RewardsRecord{
			ID:               record.Id,
			RewardsAddress:   record.RewardsAddress,
			Rewards:          wasmdTypes.NewWasmCoins(record.Rewards),
			CalculatedHeight: record.CalculatedHeight,
			CalculatedTime:   record.CalculatedTime.Format(time.RFC3339Nano),
		})
	}

	return resp
}

// Validate performs request fields validation.
func (r RewardsRecordsRequest) Validate() error {
	if _, err := sdk.AccAddressFromBech32(r.RewardsAddress); err != nil {
		return fmt.Errorf("rewardsAddress: parsing: %w", err)
	}

	return nil
}

// MustGetRewardsAddress returns the rewards address as sdk.AccAddress.
// CONTRACT: panics in case of an error (should not happen since we validate the request).
func (r RewardsRecordsRequest) MustGetRewardsAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(r.RewardsAddress)
	if err != nil {
		// Should not happen since we validate the request before this call
		panic(fmt.Errorf("wasm bindings: rewardsRecordsRequest request: parsing rewardsAddress: %w", err))
	}

	return addr
}
