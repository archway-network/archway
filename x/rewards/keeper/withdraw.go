package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/archway-network/archway/x/rewards/types"
)

// WithdrawRewardsByRecordsLimit performs the rewards distribution for the given rewards address and the number of record to use.
func (k Keeper) WithdrawRewardsByRecordsLimit(ctx sdk.Context, rewardsAddr sdk.AccAddress, recordsLimit uint64) (sdk.Coins, int, error) {
	recordsLimitMax := k.MaxWithdrawRecords(ctx)

	// Use the default limit if not specified
	if recordsLimit == 0 {
		recordsLimit = recordsLimitMax
	}

	// Msg post-validateBasic check
	if recordsLimit > recordsLimitMax {
		return nil, 0, sdkErrors.Wrapf(types.ErrInvalidRequest, "max withdraw records (%d) exceeded", recordsLimitMax)
	}

	// Get all rewards records for the given address by limit
	pageReq := &query.PageRequest{Limit: recordsLimit}
	records, _, err := k.state.RewardsRecord(ctx).GetRewardsRecordByRewardsAddressPaginated(rewardsAddr, pageReq)
	if err != nil {
		return nil, 0, sdkErrors.Wrap(types.ErrInternal, err.Error())
	}

	return k.withdrawRewardsByRecords(ctx, rewardsAddr, records), len(records), nil
}

// WithdrawRewardsByRecordIDs performs the rewards distribution for the given rewards address and record IDs.
func (k Keeper) WithdrawRewardsByRecordIDs(ctx sdk.Context, rewardsAddr sdk.AccAddress, recordIDs []uint64) (sdk.Coins, int, error) {
	// Msg post-validateBasic check
	if maxRecords := k.MaxWithdrawRecords(ctx); uint64(len(recordIDs)) > maxRecords {
		return nil, 0, sdkErrors.Wrapf(types.ErrInvalidRequest, "max withdraw records (%d) exceeded", maxRecords)
	}

	rewardsState := k.state.RewardsRecord(ctx)
	rewardsAddrStr := rewardsAddr.String()

	// Check that provided IDs do exist and belong to the given address
	records := make([]types.RewardsRecord, 0, len(recordIDs))
	for _, id := range recordIDs {
		record, found := rewardsState.GetRewardsRecord(id)
		if !found {
			return nil, 0, sdkErrors.Wrapf(types.ErrInvalidRequest, "rewards record (%d): not found", id)
		}
		if record.RewardsAddress != rewardsAddrStr {
			return nil, 0, sdkErrors.Wrapf(types.ErrInvalidRequest, "rewards record (%d): address mismatch", id)
		}

		records = append(records, record)
	}

	return k.withdrawRewardsByRecords(ctx, rewardsAddr, records), len(records), nil
}

// withdrawRewardsByRecords performs the rewards distribution for the given rewards address and records.
// Handler emits the distribution event and prunes the used records.
func (k Keeper) withdrawRewardsByRecords(ctx sdk.Context, rewardsAddr sdk.AccAddress, records []types.RewardsRecord) sdk.Coins {
	// Aggregate total rewards to distribute
	totalRewards := sdk.NewCoins()
	for _, record := range records {
		totalRewards = totalRewards.Add(record.Rewards...)
	}

	// Transfer rewards and emit distribution event
	if !totalRewards.IsZero() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ContractRewardCollector, rewardsAddr, totalRewards); err != nil {
			panic(fmt.Errorf("sending rewards (%s) to the rewards address (%s): %w", totalRewards, rewardsAddr, err))
		}

		types.EmitRewardsWithdrawEvent(ctx, rewardsAddr, totalRewards)
	}

	// Clean up (safe if there were no rewards)
	k.state.RewardsRecord(ctx).DeleteRewardsRecords(records...)

	return totalRewards
}
