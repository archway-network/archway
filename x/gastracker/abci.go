package gastracker

import (
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

type RewardTransferKeeper interface {
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
}

type MintParamsKeeper interface {
	GetParams(ctx sdk.Context) (params mintTypes.Params)
	GetMinter(ctx sdk.Context) (minter mintTypes.Minter)
}

func EmitRewardPayingEvent(context sdk.Context, rewardAddress string, rewardsPayed sdk.Coins, leftOverRewards []*sdk.DecCoin) error {
	rewards := make([]*sdk.Coin, len(rewardsPayed))
	for i := range rewards {
		rewards[i] = &rewardsPayed[i]
	}

	return context.EventManager().EmitTypedEvent(&gstTypes.RewardDistributionEvent{
		RewardAddress:   rewardAddress,
		ContractRewards: rewards,
		LeftoverRewards: leftOverRewards,
	})
}

func EmitContractRewardCalculationEvent(context sdk.Context, contractAddress string, gasConsumed uint64, inflationReward sdk.DecCoin, contractRewards sdk.DecCoins, metadata *gstTypes.ContractInstanceMetadata) error {
	rewards := make([]*sdk.DecCoin, len(contractRewards))
	for i := range rewards {
		rewards[i] = &contractRewards[i]
	}

	return context.EventManager().EmitTypedEvent(&gstTypes.ContractRewardCalculationEvent{
		ContractAddress:  contractAddress,
		GasConsumed:      gasConsumed,
		InflationRewards: &inflationReward,
		ContractRewards:  rewards,
		Metadata:         metadata,
	})
}

func BeginBlock(context sdk.Context, _ abci.RequestBeginBlock, gasTrackingKeeper GasTrackingKeeper, rewardTransferKeeper RewardTransferKeeper, mintParamsKeeper MintParamsKeeper) {
	lastBlockGasTracking := resetBlockGasTracking(context, gasTrackingKeeper)

	if !gasTrackingKeeper.IsGasTrackingEnabled(context) { // No rewards or calculations should take place
		return
	}
	context.Logger().Info("Got the tracking for block", "BlockTxDetails", lastBlockGasTracking)

	contractTotalInflationRewards := getContractInflationRewards(context, mintParamsKeeper) // 20% of the rewards distributed on every block

	totalContractRewardsPerBlock, rewardAddresses, rewardsByAddress := getContractRewards(context, lastBlockGasTracking, gasTrackingKeeper, contractTotalInflationRewards)

	// We need to commit pending metadata before we return but after we calculated rewards.
	commitPendingMetadata(context, gasTrackingKeeper)

	// Either the tx did not collect any fee or no contracts were executed
	// So, no need to continue execution
	if totalContractRewardsPerBlock == nil || totalContractRewardsPerBlock.IsZero() {
		return
	}

	totalFeeToBeCollected := make(sdk.Coins, len(totalContractRewardsPerBlock))
	for i := range totalFeeToBeCollected {
		totalFeeToBeCollected[i] = sdk.NewCoin(totalContractRewardsPerBlock[i].Denom, totalContractRewardsPerBlock[i].Amount.Ceil().RoundInt())
	}

	err := rewardTransferKeeper.SendCoinsFromModuleToModule(context, authTypes.FeeCollectorName, ContractRewardCollector, totalFeeToBeCollected)
	if err != nil {
		panic(err)
	}

	distributeRewards(context, rewardAddresses, rewardsByAddress, gasTrackingKeeper, rewardTransferKeeper)
}

func commitPendingMetadata(context sdk.Context, gasTrackingKeeper GasTrackingKeeper) {
	numberOfEntriesCommitted, err := gasTrackingKeeper.CommitPendingContractMetadata(context)
	if err != nil {
		panic(err)
	}
	context.Logger().Info("Committed pending metadata change", "NumberOfMetadataCommitted", numberOfEntriesCommitted)
}

// resetBlockGasTracking resets the current status and returns the last blockGasTracking
func resetBlockGasTracking(context sdk.Context, gasTrackingKeeper GasTrackingKeeper) gstTypes.BlockGasTracking {
	lastBlockGasTracking := getCurrentBlockGasTracking(context, gasTrackingKeeper)

	if err := gasTrackingKeeper.TrackNewBlock(context); err != nil {
		panic(err)
	}
	return lastBlockGasTracking
}

// getCurrentBlockGasTracking returns the actual block gas tracking, panics if empty and block height is bigger than one.
func getCurrentBlockGasTracking(context sdk.Context, gasTrackingKeeper GasTrackingKeeper) gstTypes.BlockGasTracking {
	currentBlockTrackingInfo, err := gasTrackingKeeper.GetCurrentBlockTracking(context)
	if err != nil {
		switch err {
		case gstTypes.ErrBlockTrackingDataNotFound:
			// Only panic when there was a previous block
			if context.BlockHeight() > 1 {
				panic(err)
			}
		default:
			panic(err)
		}
	}

	return currentBlockTrackingInfo
}

// distributeRewards distributes the calculated rewards to all the contracts owners.
func distributeRewards(context sdk.Context, rewardAddresses []string, rewardsByAddress map[string]sdk.DecCoins, gasTrackingKeeper GasTrackingKeeper, rewardTransferKeeper RewardTransferKeeper) {
	for _, rewardAddressStr := range rewardAddresses {
		rewardAddress, err := sdk.AccAddressFromBech32(rewardAddressStr)
		if err != nil {
			panic(err)
		}

		rewards := rewardsByAddress[rewardAddressStr]
		// TODO: We should take leftOverThreshold from governance
		rewardsToBePayed, err := gasTrackingKeeper.CreateOrMergeLeftOverRewardEntry(context, rewardAddress, rewards, 1)
		if err != nil {
			panic(err)
		}

		err = rewardTransferKeeper.SendCoinsFromModuleToAccount(context, ContractRewardCollector, rewardAddress, rewardsToBePayed)
		if err != nil {
			panic(err)
		}

		leftOverEntry, err := gasTrackingKeeper.GetLeftOverRewardEntry(context, rewardAddress)
		if err != nil {
			panic(err)
		}

		err = EmitRewardPayingEvent(context, rewardAddressStr, rewardsToBePayed, leftOverEntry.ContractRewards)
		if err != nil {
			panic(err)
		}

		context.Logger().Info("Reward allocation details:", "rewardPayed", rewardsToBePayed, "leftOverEntry", leftOverEntry.ContractRewards)
	}
}

// getContractRewards returns the total rewards and the rewards per contract based on the calculations.
func getContractRewards(context sdk.Context, blockGasTracking gstTypes.BlockGasTracking, gasTrackingKeeper GasTrackingKeeper, contractTotalInflationRewards sdk.DecCoin) (sdk.DecCoins, []string, map[string]sdk.DecCoins) {
	// To enforce a map iteration order. This isn't strictly necessary but is only
	// done to make this code more deterministic.
	rewardAddresses := make([]string, 0)
	rewardsByAddress := make(map[string]sdk.DecCoins)

	totalContractRewardsPerBlock := make(sdk.DecCoins, 0)
	for _, txTrackingInfo := range blockGasTracking.TxTrackingInfos {
		// We generate empty coins based on the fees coins.
		totalContractRewardsInTx := make(sdk.DecCoins, len(txTrackingInfo.MaxContractRewards))
		for i, _ := range totalContractRewardsInTx {
			totalContractRewardsInTx[i] = sdk.NewDecCoin(txTrackingInfo.MaxContractRewards[i].Denom, sdk.NewInt(0))
		}

		for _, contractTrackingInfo := range txTrackingInfo.ContractTrackingInfos {
			var contractInflationReward sdk.DecCoin
			contractAddress, err := sdk.AccAddressFromBech32(contractTrackingInfo.Address)
			if err != nil {
				panic(err)
			}

			metadata, err := gasTrackingKeeper.GetContractMetadata(context, contractAddress)
			if err != nil {
				panic(err)
			}
			context.Logger().Info("Got the metadata", "Metadata", metadata)

			contractRewards := make(sdk.DecCoins, 0, 0)

			gasConsumedInContract := sdk.NewDecFromBigInt(ConvertUint64ToBigInt(contractTrackingInfo.GasConsumed))

			if gasTrackingKeeper.IsDappInflationRewardsEnabled(context) && context.BlockGasMeter().Limit() > 0 {
				blockGasLimit := sdk.NewDecFromBigInt(ConvertUint64ToBigInt(context.BlockGasMeter().Limit()))
				contractInflationReward = sdk.NewDecCoinFromDec(contractTotalInflationRewards.Denom, contractTotalInflationRewards.Amount.Mul(gasConsumedInContract).Quo(blockGasLimit))
				context.Logger().Info("Calculated contract inflation rewards:", "contractAddress", contractAddress, "contractInflationReward", contractInflationReward)
				contractRewards = contractRewards.Add(contractInflationReward)
			}

			if !gasTrackingKeeper.IsGasRebateToUserEnabled(context) || !metadata.GasRebateToUser {
				maxGasAllowedInTx := sdk.NewDecFromBigInt(ConvertUint64ToBigInt(txTrackingInfo.MaxGasAllowed))

				// Calc premium fees
				var gasUsageForUsageRewards = gasConsumedInContract
				if metadata.CollectPremium && gasTrackingKeeper.IsContractPremiumEnabled(context) {
					premiumGas := gasConsumedInContract.
						Mul(sdk.NewDecFromBigInt(ConvertUint64ToBigInt(metadata.PremiumPercentageCharged))).
						QuoInt64(100)
					gasUsageForUsageRewards = gasUsageForUsageRewards.Add(premiumGas)
				}

				if gasTrackingKeeper.IsGasRebateToContractEnabled(context) {
					for _, rewardCoin := range txTrackingInfo.MaxContractRewards {
						contractRewards = contractRewards.Add(sdk.NewDecCoinFromDec(
							rewardCoin.Denom, rewardCoin.Amount.Mul(gasUsageForUsageRewards).Quo(maxGasAllowedInTx)))
					}
					context.Logger().
						Info("Calculated contract gas rebate rewards:",
							"contractAddress", contractAddress, "contractGasReward", contractRewards)
				}
			} else {
				context.Logger().Info("Contract is not eligible for gas rewards, skipping calculation.", "contractAddress", contractAddress)
			}

			if _, ok := rewardsByAddress[metadata.RewardAddress]; !ok {
				rewardAddresses = append(rewardAddresses, metadata.RewardAddress)
				rewardsByAddress[metadata.RewardAddress] = contractRewards
			} else {
				rewardsByAddress[metadata.RewardAddress] = rewardsByAddress[metadata.RewardAddress].Add(contractRewards...)
			}

			totalContractRewardsInTx = totalContractRewardsInTx.Add(contractRewards...)

			if err = EmitContractRewardCalculationEvent(context, contractAddress.String(), contractTrackingInfo.GasConsumed, contractInflationReward, contractRewards, &metadata); err != nil {
				panic(err)
			}

			context.Logger().Info("Calculated Contract rewards:", "contractAddress", contractAddress, "contractRewards", contractRewards)
		}

		totalContractRewardsPerBlock = totalContractRewardsPerBlock.Add(totalContractRewardsInTx...)
	}

	return totalContractRewardsPerBlock, rewardAddresses, rewardsByAddress
}

// getContractInflationRewards returns the percentage of the block rewards that are dedicated to contracts
// TODO now is 20% of the block rewards hardcoded.
func getContractInflationRewards(context sdk.Context, mintParamsKeeper MintParamsKeeper) sdk.DecCoin {
	totalInflationRatePerBlock := getInflationFeeForLastBlock(context, mintParamsKeeper)

	// TODO: Take the percentage value from governance
	contractTotalInflationRewards := sdk.NewDecCoinFromDec(totalInflationRatePerBlock.Denom, totalInflationRatePerBlock.Amount.MulInt64(20).QuoInt64(100))

	return contractTotalInflationRewards
}

// getInflationFeeForLastBlock returns the inflation per block. (Annual Inflation / NumblocksPerYear)
func getInflationFeeForLastBlock(context sdk.Context, mintParamsKeeper MintParamsKeeper) sdk.DecCoin {
	minter := mintParamsKeeper.GetMinter(context)
	params := mintParamsKeeper.GetParams(context)
	totalInflationFee := sdk.NewDecCoinFromCoin(minter.BlockProvision(params))

	return totalInflationFee
}
