package keeper

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/pkg"
	"github.com/archway-network/archway/x/rewards/types"
)

type (
	// blockRewardsDistributionState is used to gather gas usage and rewards for a block on a contract basis.
	blockRewardsDistributionState struct {
		Height             int64                                        // block height
		Txs                map[uint64]uint64                            // gas usage per transaction [key: txID, value: total gas]
		Contracts          map[string]*contractRewardsDistributionState // contract rewards state [key: contract address]
		RewardsTotal       sdk.Coins                                    // total rewards for the block (inflationary + txs rewards)
		RewardsDistributed sdk.Coins                                    // total rewards distributed for the block
	}

	// contractRewardsDistributionState is used to gather gas usage and rewards for a contract.
	contractRewardsDistributionState struct {
		ContractAddress sdk.AccAddress          // contract address
		Metadata        *types.ContractMetadata // metadata for this contract (might be nil if not set)

		BlockGasUsed uint64            // total gas used in the block (all operations across all transaction)
		TxGasUsed    map[uint64]uint64 // total gas used in a transaction (all operations across one transaction) [key: txID, value: gas used]

		FeeRewards          sdk.Coins // fee rewards for this contract (for all txs)
		InflationaryRewards sdk.Coin  // inflation rewards for this contract (for the block)
	}
)

// AllocateBlockRewards creates rewards records for the given block height.
func (k Keeper) AllocateBlockRewards(ctx sdk.Context, height int64) {
	blockDistrState := k.estimateBlockGasUsage(ctx, height)
	blockDistrState = k.estimateBlockRewards(ctx, blockDistrState)
	k.createRewardsRecords(ctx, blockDistrState)
	k.cleanupRewardsPool(ctx, blockDistrState)
	k.cleanupTracking(ctx, height)
}

// estimateBlockGasUsage creates a new distribution state for the given block height.
// Func iterates over all tracked transactions and estimates gas usage for each contract (on block and tx levels) merging operations.
func (k Keeper) estimateBlockGasUsage(ctx sdk.Context, height int64) *blockRewardsDistributionState {
	metadataState := k.state.ContractMetadataState(ctx)

	// Get all tracked transactions by the x/tracking module
	blockGasTrackingInfo := k.trackingKeeper.GetBlockTrackingInfo(ctx, height)

	// Create a new block rewards distribution state and fill it up
	blockDistrState := &blockRewardsDistributionState{
		Height:             height,
		Txs:                make(map[uint64]uint64, len(blockGasTrackingInfo.Txs)),
		Contracts:          make(map[string]*contractRewardsDistributionState, 0),
		RewardsTotal:       sdk.NewCoins(),
		RewardsDistributed: sdk.NewCoins(),
	}

	// Fill up gas usage iterating over all tracked transactions and contract operations
	for _, txGasTrackingInfo := range blockGasTrackingInfo.Txs {
		// Set total gas used by the transaction
		blockDistrState.Txs[txGasTrackingInfo.Info.Id] = txGasTrackingInfo.Info.TotalGas

		// Estimate contract operations total gas used for this transaction
		for _, contractOp := range txGasTrackingInfo.ContractOperations {
			opGasUsed, opEligible := contractOp.GasUsed()
			if !opEligible {
				// Skip noop operation (should not happen since we're tracking an actual WASM usage)
				k.Logger(ctx).Debug("Noop contract operation found (skip)", "txID", contractOp.TxId, "opID", contractOp.Id)
				continue
			}

			// Create a new contract rewards distribution state
			contractDistrState := blockDistrState.Contracts[contractOp.ContractAddress]
			if contractDistrState == nil {
				contractDistrState = &contractRewardsDistributionState{
					ContractAddress:     contractOp.MustGetContractAddress(),
					TxGasUsed:           make(map[uint64]uint64, 0),
					InflationaryRewards: sdk.Coin{Amount: sdk.ZeroInt()}, // necessary to avoid nil pointer panic on Coins.Add call
				}
				if metadata, found := metadataState.GetContractMetadata(contractDistrState.ContractAddress); found {
					contractDistrState.Metadata = &metadata
				}
				blockDistrState.Contracts[contractOp.ContractAddress] = contractDistrState
			}

			// Increase block gas usage
			contractDistrState.BlockGasUsed += opGasUsed

			// Increase tx gas usage
			txGasUsed := contractDistrState.TxGasUsed[contractOp.TxId] // 0 if not initialized
			contractDistrState.TxGasUsed[contractOp.TxId] = txGasUsed + opGasUsed
		}
	}

	return blockDistrState
}

// estimateBlockRewards update block distribution state with tracked rewards calculating reward shares per contract.
// Func iterates over all tracked transactions and estimates inflation (on block level) and fee rebate (merging
// tokens for each transaction contract has operation at) rewards for each contract.
func (k Keeper) estimateBlockRewards(ctx sdk.Context, blockDistrState *blockRewardsDistributionState) *blockRewardsDistributionState {
	txRewardsState := k.state.TxRewardsState(ctx)

	// Fetch tracked block rewards by the x/rewards module (might not be found in case this reward is disabled)
	inlfationRewardsEligible := false
	blockRewards, found := k.state.BlockRewardsState(ctx).GetBlockRewards(blockDistrState.Height)
	if found && blockRewards.HasRewards() {
		blockDistrState.RewardsTotal = blockDistrState.RewardsTotal.Add(blockRewards.InflationRewards)
		if blockRewards.HasGasLimit() {
			inlfationRewardsEligible = true
		}
	} else {
		k.Logger(ctx).Debug("No inflation rewards to distribute (no record / empty coin / gas limit not set)", "height", blockDistrState.Height)
	}

	// Fetch tracked transactions rewards by the x/rewards module (some might not be found in case this reward is disabled)
	txsRewards := make(map[uint64]sdk.Coins, len(blockDistrState.Txs))
	for txID := range blockDistrState.Txs {
		txRewards, found := txRewardsState.GetTxRewards(txID)
		if found && txRewards.HasRewards() {
			txsRewards[txID] = txRewards.FeeRewards
			blockDistrState.RewardsTotal = blockDistrState.RewardsTotal.Add(txRewards.FeeRewards...)
		} else {
			k.Logger(ctx).Debug("No tx fee rebate rewards to distribute (no record / empty coins)", "txID", txID)
		}
	}

	// Estimate contract rewards
	for _, contractDistrState := range blockDistrState.Contracts {
		// Estimate contract inflation rewards
		if inlfationRewardsEligible {
			gasUsed := pkg.NewDecFromUint64(contractDistrState.BlockGasUsed)
			rewardsShare := gasUsed.Quo(pkg.NewDecFromUint64(blockRewards.MaxGas))

			inflationRewards := sdk.NewCoin(
				blockRewards.InflationRewards.Denom,
				blockRewards.InflationRewards.Amount.ToDec().Mul(rewardsShare).TruncateInt(),
			)
			contractDistrState.InflationaryRewards = inflationRewards
		}

		// Estimate contract tx fee rebate rewards (sum of all transactions involved)
		for txID, gasUsed := range contractDistrState.TxGasUsed {
			txFees, feeRewardsEligible := txsRewards[txID]
			if !feeRewardsEligible {
				continue
			}

			gasTotal := pkg.NewDecFromUint64(blockDistrState.Txs[txID])
			rewardsShare := pkg.NewDecFromUint64(gasUsed).Quo(gasTotal)

			for _, feeCoin := range txFees {
				feeRewards := sdk.NewCoin(
					feeCoin.Denom,
					feeCoin.Amount.ToDec().Mul(rewardsShare).TruncateInt(),
				)
				contractDistrState.FeeRewards = contractDistrState.FeeRewards.Add(feeRewards)
			}
		}
	}

	return blockDistrState
}

// createRewardsRecords creates types.RewardsRecord entries for a respective reward addresses if set (otherwise, skip)
// and emit calculation events. An actual distribution (x/bank transfer) is performed later.
// Leftovers caused by Int truncation or by a tx-less block (inflation rewards are tracked even if there were no transactions)
// stay in the pool.
func (k Keeper) createRewardsRecords(ctx sdk.Context, blockDistrState *blockRewardsDistributionState) {
	rewardsRecordState := k.state.RewardsRecord(ctx)
	calculationHeight, calculationTime := ctx.BlockHeight(), ctx.BlockTime()

	// Convert contract distribution states to a sorted slice preventing the consensus failure due to x/bank operations order.
	// Filter out contracts without: rewards, metadata or rewardsAddress.
	// Emit calculation events for each contract.
	contractStates := make([]*contractRewardsDistributionState, 0, len(blockDistrState.Contracts))
	for _, contractDistrState := range blockDistrState.Contracts {
		// Emit calculation event
		types.EmitContractRewardCalculationEvent(
			ctx,
			contractDistrState.ContractAddress,
			contractDistrState.BlockGasUsed,
			contractDistrState.InflationaryRewards,
			contractDistrState.FeeRewards,
			contractDistrState.Metadata,
		)

		// Filter out
		if contractDistrState.FeeRewards.IsZero() && contractDistrState.InflationaryRewards.IsZero() {
			k.Logger(ctx).Debug("No contract rewards to distribute (skip)", "contract", contractDistrState.ContractAddress)
			continue
		}
		if contractDistrState.Metadata == nil {
			k.Logger(ctx).Debug("Contract metadata is not set (skip)", "contract", contractDistrState.ContractAddress)
			continue
		}
		if !contractDistrState.Metadata.HasRewardsAddress() {
			k.Logger(ctx).Debug("Contract rewards address is not set (skip)", "contract", contractDistrState.ContractAddress)
			continue
		}

		contractStates = append(contractStates, contractDistrState)
	}
	sort.Slice(contractStates, func(i, j int) bool {
		return contractStates[i].ContractAddress.String() < contractStates[j].ContractAddress.String()
	})

	// Distribute
	for _, contractDistrState := range contractStates {
		// Transfer to the rewardsAddress
		rewardsAddr := contractDistrState.Metadata.MustGetRewardsAddress()
		rewards := sdk.NewCoins().
			Add(contractDistrState.InflationaryRewards).
			Add(contractDistrState.FeeRewards...)

		// Create a new record
		rewardsRecordState.CreateRewardsRecord(rewardsAddr, rewards, calculationHeight, calculationTime)

		// Update the total rewards distributed counter
		blockDistrState.RewardsDistributed = blockDistrState.RewardsDistributed.Add(rewards...)
	}
}

// cleanupTracking prunes all tracking data for the given block height for x/tracking and x/rewards modules.
func (k Keeper) cleanupTracking(ctx sdk.Context, height int64) {
	// We can prune the previous block ({height}), but that makes tracking CLI queries useless as there won't be any data.
	// Pruning history block also makes e2e tests possible.
	heightToPrune := height - 10
	if heightToPrune <= 0 {
		return
	}

	k.trackingKeeper.RemoveBlockTrackingInfo(ctx, heightToPrune)
	k.state.DeleteBlockRewardsCascade(ctx, heightToPrune)
}

// cleanupRewardsPool transfers all undistributed block rewards to the treasury pool.
func (k Keeper) cleanupRewardsPool(ctx sdk.Context, blockDistrState *blockRewardsDistributionState) {
	rewardsLeftovers := blockDistrState.RewardsTotal.Sub(blockDistrState.RewardsDistributed)
	if rewardsLeftovers.Empty() {
		return
	}

	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ContractRewardCollector, types.TreasuryCollector, rewardsLeftovers); err != nil {
		panic(fmt.Errorf("failed to transfer undistributed rewards (%s) to %s: %w", rewardsLeftovers, types.TreasuryCollector, err))
	}
}
