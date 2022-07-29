package keeper

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/pkg"
	"github.com/archway-network/archway/x/rewards/types"
)

type (
	// blockRewardsDistributionState is used to gather gas usage and rewards for a block.
	blockRewardsDistributionState struct {
		Height  int64   // block height
		GasUsed sdk.Dec // total gas used by the block
		Txs     []*txRewardsDistributionState
	}

	// txRewardsDistributionState is used to gather gas usage and rewards for a transaction.
	txRewardsDistributionState struct {
		TxID      uint64  // transaction ID (x/tracking unique ID)
		GasUsed   sdk.Dec // total gas used by the transaction
		Contracts []*contractRewardsDistributionState
	}

	// contractRewardsDistributionState is used to gather gas usage and rewards for a contract (merge contract operations data).
	contractRewardsDistributionState struct {
		ContractAddress     sdk.AccAddress          // contract address
		GasUsed             sdk.Dec                 // total gas used by all the contract operations
		FeeRewards          sdk.Coins               // fee rewards for this contract
		InflationaryRewards sdk.Coin                // inflation rewards for this contract
		Metadata            *types.ContractMetadata // metadata for this contract (might be nil if not set)
	}
)

// DistributeRewards distributes rewards for the given block height.
func (k Keeper) DistributeRewards(ctx sdk.Context, height int64) {
	blockDistrState := k.estimateBlockGasUsage(ctx, height)
	blockDistrState = k.estimateBlockRewards(ctx, blockDistrState)
	k.distributeBlockRewards(ctx, blockDistrState)
}

// estimateBlockGasUsage creates a new distribution state for the given block height.
// Func iterates over all tracked transactions and estimates gas usage for: block, txs and contracts.
func (k Keeper) estimateBlockGasUsage(ctx sdk.Context, height int64) *blockRewardsDistributionState {
	metadataState := k.state.ContractMetadataState(ctx)

	// Create a new block rewards distribution state and fill it up
	blockDistrState := &blockRewardsDistributionState{
		Height:  height,
		GasUsed: sdk.ZeroDec(),
	}

	// Get all tracked transactions by the x/tracking module
	blockGasTrackingInfo := k.trackingView.GetBlockTrackingInfo(ctx, height)
	blockDistrState.Txs = make([]*txRewardsDistributionState, 0, len(blockGasTrackingInfo.Txs))

	// Fill up gas usage per transaction
	for _, txGasTrackingInfo := range blockGasTrackingInfo.Txs {
		// Skip noop transaction (tx rewards will stay in the pool)
		if !txGasTrackingInfo.Info.HasGasUsage() {
			continue
		}

		// Estimate contract operations total gas used (could be multiple ops per contract, so we merge them)
		contractDistrStatesSet := make(map[string]*contractRewardsDistributionState, len(txGasTrackingInfo.ContractOperations))
		for _, contractOp := range txGasTrackingInfo.ContractOperations {
			opGasUsed, opEligible := contractOp.GasUsed()
			if !opEligible {
				// Skip noop operation (should not happen since we're tracking an actual WASM usage)
				k.Logger(ctx).Debug("Noop contract operation found", "txID", contractOp.TxId, "opID", contractOp.Id)
				continue
			}

			contractDistrState := contractDistrStatesSet[contractOp.ContractAddress]
			if contractDistrState == nil {
				contractDistrState = &contractRewardsDistributionState{
					ContractAddress: contractOp.MustGetContractAddress(),
					GasUsed:         sdk.ZeroDec(),
				}
				if metadata, found := metadataState.GetContractMetadata(contractDistrState.ContractAddress); found {
					contractDistrState.Metadata = &metadata
				}

				contractDistrStatesSet[contractOp.ContractAddress] = contractDistrState
			}
			contractDistrState.GasUsed = contractDistrState.GasUsed.Add(pkg.NewDecFromUint64(opGasUsed))
		}

		// Create a new tx rewards distribution state and fill it up
		// We sort the operations slice to prevent the consensus failure due to the order of operations
		txDistState := &txRewardsDistributionState{
			TxID:      txGasTrackingInfo.Info.Id,
			GasUsed:   pkg.NewDecFromUint64(txGasTrackingInfo.Info.TotalGas),
			Contracts: make([]*contractRewardsDistributionState, 0, len(contractDistrStatesSet)),
		}
		for _, contractRewardsState := range contractDistrStatesSet {
			txDistState.Contracts = append(txDistState.Contracts, contractRewardsState)
		}
		sort.Slice(txDistState.Contracts, func(i, j int) bool {
			return txDistState.Contracts[i].ContractAddress.String() < txDistState.Contracts[j].ContractAddress.String()
		})

		// Append tx distr state updating the block gas used
		blockDistrState.GasUsed = blockDistrState.GasUsed.Add(txDistState.GasUsed)
		blockDistrState.Txs = append(blockDistrState.Txs, txDistState)
	}

	return blockDistrState
}

// estimateBlockRewards update block distribution state with tracked rewards calculating reward shares per contract.
// Func iterates over all tracked transactions and estimates rewards for each contract.
func (k Keeper) estimateBlockRewards(ctx sdk.Context, blockDistrState *blockRewardsDistributionState) *blockRewardsDistributionState {
	txRewardsState := k.state.TxRewardsState(ctx)

	// Get tracked block rewards by the x/rewards module (might not be found in case this reward is disabled)
	blockRewards, blockRewardsFound := k.state.BlockRewardsState(ctx).GetBlockRewards(blockDistrState.Height)

	// Estimate reward shares for each contract operation
	for _, txDistrState := range blockDistrState.Txs {
		// Get tracked transaction rewards by the x/rewards module (might not be found in case this reward is disabled)
		txRewards, txRewardsFound := txRewardsState.GetTxRewards(txDistrState.TxID)

		for _, contractDistrState := range txDistrState.Contracts {
			// Estimate contract fee rewards
			if txRewardsFound && txRewards.HasRewards() {
				rewardsShare := contractDistrState.GasUsed.Quo(txDistrState.GasUsed)

				rewardCoins := sdk.NewCoins()
				for _, feeCoin := range txRewards.FeeRewards {
					rewardCoins = rewardCoins.Add(sdk.NewCoin(
						feeCoin.Denom,
						feeCoin.Amount.ToDec().Mul(rewardsShare).TruncateInt(),
					))
				}
				contractDistrState.FeeRewards = rewardCoins
			}

			// Estimate contract inflation rewards
			if blockRewardsFound && blockRewards.HasRewards() {
				rewardsShare := contractDistrState.GasUsed.Quo(blockDistrState.GasUsed)

				contractDistrState.InflationaryRewards = sdk.NewCoin(
					blockRewards.InflationRewards.Denom,
					blockRewards.InflationRewards.Amount.ToDec().Mul(rewardsShare).TruncateInt(),
				)
			}
		}
	}

	return blockDistrState
}

// distributeBlockRewards distributes block rewards to respective reward addresses if set (otherwise, skip) and emit events.
// Func sends rewards to the respective reward addresses (is set) and emits events.
// Leftovers caused by Int truncation or by a tx-less block (inflation rewards are tracked even
// if there were no transactions) stay in the pool.
func (k Keeper) distributeBlockRewards(ctx sdk.Context, blockDistrState *blockRewardsDistributionState) {
	for _, txDistrState := range blockDistrState.Txs {
		for _, contractDistrState := range txDistrState.Contracts {
			// Emit calculation event
			types.EmitContractRewardCalculationEvent(
				ctx,
				contractDistrState.ContractAddress,
				uint64(contractDistrState.GasUsed.TruncateInt64()),
				contractDistrState.InflationaryRewards,
				contractDistrState.FeeRewards,
				contractDistrState.Metadata,
			)

			// Skip cases
			if contractDistrState.Metadata == nil {
				k.Logger(ctx).Debug("Contract rewards distribution skipped (no metadata found)", "contractAddress", contractDistrState.ContractAddress)
				continue
			}
			if !contractDistrState.Metadata.HasRewardsAddress() {
				k.Logger(ctx).Debug("Contract rewards distribution skipped (rewards address not set)", "contractAddress", contractDistrState.ContractAddress)
				continue
			}
			rewardsAddr := contractDistrState.Metadata.MustGetRewardsAddress()

			// Distribute
			rewards := sdk.NewCoins(contractDistrState.FeeRewards...).Add(contractDistrState.InflationaryRewards)
			if rewards.IsZero() {
				k.Logger(ctx).Debug("Contract rewards distribution skipped (no rewards)", "contractAddress", contractDistrState.ContractAddress)
				continue
			}

			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ContractRewardCollector, rewardsAddr, rewards); err != nil {
				panic(fmt.Errorf("sending rewards (%s) to rewards address (%s) for the contract (%s): %w", contractDistrState.FeeRewards, rewardsAddr, contractDistrState.ContractAddress, err))
			}

			// Emit distribution event
			types.EmitContractRewardDistributionEvent(
				ctx,
				contractDistrState.ContractAddress,
				rewardsAddr,
				rewards,
			)
		}
	}
}
