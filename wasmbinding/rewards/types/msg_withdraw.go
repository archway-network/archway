package types

import (
	"fmt"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmVmTypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// WithdrawRewardsRequest is the Msg.WithdrawRewards request.
type WithdrawRewardsRequest struct {
	// RecordsLimit defines the maximum number of RewardsRecord objects to process.
	// Limit should not exceed the MaxWithdrawRecords param value.
	// If 0 value is passed, the MaxWithdrawRecords value is used.
	// Only one of (RecordsLimit, RecordIDs) should be set.
	RecordsLimit *uint64 `json:"records_limit"`
	// RecordIDs defines specific RewardsRecord object IDs to process.
	// Only one of (RecordsLimit, RecordIDs) should be set.
	RecordIDs []uint64 `json:"record_ids"`
}

// WithdrawRewardsResponse is the Msg.WithdrawRewards response.
type WithdrawRewardsResponse struct {
	// RecordsNum is the number of RewardsRecord objects processed by the request.
	RecordsNum uint64 `json:"records_num"`
	// TotalRewards are the total rewards distributed.
	TotalRewards wasmVmTypes.Coins `json:"total_rewards"`
}

// Validate performs request fields validation.
func (r WithdrawRewardsRequest) Validate() error {
	if (r.RecordsLimit == nil && len(r.RecordIDs) == 0) || (r.RecordsLimit != nil && len(r.RecordIDs) > 0) {
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
