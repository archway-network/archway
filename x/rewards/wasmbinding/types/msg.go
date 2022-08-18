package types

import (
	"fmt"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmVmTypes "github.com/CosmWasm/wasmvm/types"
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
	WithdrawRewardsRequest struct {
		// RecordsLimit defines the maximum number of RewardsRecord objects to process.
		// Limit should not exceed the MaxWithdrawRecords param value.
		// Only one of (RecordsLimit, RecordIDs) should be set.
		RecordsLimit uint64 `json:"records_limit"`
		// RecordIDs defines specific RewardsRecord object IDs to process.
		// Only one of (RecordsLimit, RecordIDs) should be set.
		RecordIDs []uint64 `json:"record_ids"`
	}

	// WithdrawRewardsResponse is the Msg.WithdrawRewards response.
	WithdrawRewardsResponse struct {
		// RecordsNum is the number of RewardsRecord objects processed by the request.
		RecordsNum uint64 `json:"records_num"`
		// TotalRewards are the total rewards distributed.
		TotalRewards wasmVmTypes.Coins `json:"total_rewards"`
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

// ToSDK convert the UpdateMetadataRequest to a rewardsTypes.Metadata.
func (r UpdateMetadataRequest) ToSDK() rewardsTypes.ContractMetadata {
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

// Validate performs request fields validation.
func (r WithdrawRewardsRequest) Validate() error {
	if (r.RecordsLimit == 0 && len(r.RecordIDs) == 0) || (r.RecordsLimit > 0 && len(r.RecordIDs) > 0) {
		return fmt.Errorf("one of (RecordsLimit, RecordIDs) fields must be set")
	}

	idsSet := make(map[uint64]struct{})
	for _, id := range r.RecordIDs {
		if id == 0 {
			return fmt.Errorf("recordIDs: ID must be GT 0")
		}

		if _, ok := idsSet[id]; ok {
			return fmt.Errorf("recordIDs: duplicate ID (%d)", id)
		}
		idsSet[id] = struct{}{}
	}

	return nil
}

// NewWithdrawRewardsResponse creates a new WithdrawRewardsResponse.
func NewWithdrawRewardsResponse(totalRewards sdk.Coins, recordsUsed int) WithdrawRewardsResponse {
	return WithdrawRewardsResponse{
		RecordsNum:   uint64(recordsUsed),
		TotalRewards: wasmdTypes.NewWasmCoins(totalRewards),
	}
}
