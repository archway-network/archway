package module

import (
	gstTypes "github.com/archway-network/archway/x/gastracker"
	gstKeeper "github.com/archway-network/archway/x/gastracker/keeper"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"time"
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

func EmitContractRewardCalculationEvent(context sdk.Context, contractAddress string, gasConsumed sdk.Dec, inflationReward sdk.DecCoins, contractReward sdk.DecCoins, metadata *gstTypes.ContractInstanceMetadata) error {
	contractRewardEventData := make([]*sdk.DecCoin, len(contractReward))
	for i := range contractRewardEventData {
		contractRewardEventData[i] = &contractReward[i]
	}

	inflationRewardEventData := make([]*sdk.DecCoin, len(inflationReward))
	for i := range inflationReward {
		inflationRewardEventData[i] = &inflationReward[i]
	}

	return context.EventManager().EmitTypedEvent(&gstTypes.ContractRewardCalculationEvent{
		ContractAddress:  contractAddress,
		GasConsumed:      gasConsumed.RoundInt().Uint64(),
		InflationRewards: inflationRewardEventData,
		ContractRewards:  contractRewardEventData,
		Metadata:         metadata,
	})
}

func BeginBlock(context sdk.Context, _ abci.RequestBeginBlock, gasTrackingKeeper gstKeeper.GasTrackingKeeper, rewardTransferKeeper RewardTransferKeeper, mintParamsKeeper MintParamsKeeper) {
	defer telemetry.ModuleMeasureSince(gstTypes.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	lastBlockGasTracking := resetBlockGasTracking(context, gasTrackingKeeper)

	if !gasTrackingKeeper.IsGasTrackingEnabled(context) { // No rewards or calculations should take place
		return
	}
	context.Logger().Debug("Got the tracking for block", "BlockTxDetails", lastBlockGasTracking)

	contractTotalInflationRewards := getContractInflationRewards(context, mintParamsKeeper) // 20% of the rewards distributed on every block

	contractRewardData := getContractRewards(context, lastBlockGasTracking, gasTrackingKeeper, contractTotalInflationRewards)
	totalGasRewardsPerBlock := contractRewardData.totalGasRewards
	rewardAddresses := contractRewardData.rewardAddresses
	gasRewardByAddress := contractRewardData.gasRewardByRewardAddress
	totalInflationRewardsPerBlock := contractRewardData.totalInflationRewards
	contractAddresses := contractRewardData.contractAddresses
	inflationRewardByContractAddress := contractRewardData.inflationRewardByContractAddress

	// Either the tx did not collect any fee or no contracts were executed
	// So, no need to continue execution
	if totalGasRewardsPerBlock != nil && !totalGasRewardsPerBlock.IsZero() {
		totalGasRewardToBeCollected := make(sdk.Coins, len(totalGasRewardsPerBlock))
		for i := range totalGasRewardToBeCollected {
			totalGasRewardToBeCollected[i] = sdk.NewCoin(totalGasRewardsPerBlock[i].Denom, totalGasRewardsPerBlock[i].Amount.Ceil().RoundInt())
		}

		err := rewardTransferKeeper.SendCoinsFromModuleToModule(context, authTypes.FeeCollectorName, gstTypes.GasRewardCollector, totalGasRewardToBeCollected)
		if err != nil {
			panic(err)
		}

		distributeGasRewards(context, rewardAddresses, gasRewardByAddress, gasTrackingKeeper, rewardTransferKeeper)
	}

	if totalInflationRewardsPerBlock != nil && !totalInflationRewardsPerBlock.IsZero() {
		totalInflationRewardToBeCollected := make(sdk.Coins, len(totalInflationRewardsPerBlock))
		for i := range totalInflationRewardToBeCollected {
			totalInflationRewardToBeCollected[i] = sdk.NewCoin(totalInflationRewardsPerBlock[i].Denom, totalInflationRewardsPerBlock[i].Amount.Ceil().RoundInt())
		}

		err := rewardTransferKeeper.SendCoinsFromModuleToModule(context, authTypes.FeeCollectorName, gstTypes.InflationRewardAccumulator, totalInflationRewardToBeCollected)
		if err != nil {
			panic(err)
		}

		recordInflationRewards(context, gasTrackingKeeper, contractAddresses, inflationRewardByContractAddress)
	}

	// We need to commit pending metadata before we return but after we calculated rewards.
	commitPendingMetadata(context, gasTrackingKeeper)
}

func commitPendingMetadata(context sdk.Context, gasTrackingKeeper gstKeeper.GasTrackingKeeper) {
	numberOfEntriesCommitted, err := gasTrackingKeeper.CommitPendingContractMetadata(context)
	if err != nil {
		panic(err)
	}
	context.Logger().Debug("Committed pending metadata change", "NumberOfMetadataCommitted", numberOfEntriesCommitted)
}

// resetBlockGasTracking resets the current status and returns the last blockGasTracking
func resetBlockGasTracking(context sdk.Context, gasTrackingKeeper gstKeeper.GasTrackingKeeper) gstTypes.BlockGasTracking {
	lastBlockGasTracking := getCurrentBlockGasTracking(context, gasTrackingKeeper)

	if err := gasTrackingKeeper.TrackNewBlock(context); err != nil {
		panic(err)
	}
	return lastBlockGasTracking
}

// getCurrentBlockGasTracking returns the actual block gas tracking, panics if empty and block height is bigger than one.
func getCurrentBlockGasTracking(context sdk.Context, gasTrackingKeeper gstKeeper.GasTrackingKeeper) gstTypes.BlockGasTracking {
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

func recordInflationRewards(context sdk.Context, gasTrackingKeeper gstKeeper.GasTrackingKeeper, contractAddresses []string, inflationRewardByContractAddress map[string]sdk.DecCoins) {
	for _, contractAddress := range contractAddresses {
		contractAddr, err := sdk.AccAddressFromBech32(contractAddress)
		if err != nil {
			panic(err)
		}
		systemMetadata, err := gasTrackingKeeper.GetContractSystemMetadata(context, contractAddr)
		if err != nil {
			panic(err)
		}

		currentBalance := make(sdk.DecCoins, len(systemMetadata.InflationBalance))
		for i := range currentBalance {
			currentBalance[i] = *systemMetadata.InflationBalance[i]
		}

		updatedBalance := currentBalance.Add(inflationRewardByContractAddress[contractAddress]...)
		toStoreBalance := make([]*sdk.DecCoin, len(updatedBalance))
		for i := range toStoreBalance {
			toStoreBalance[i] = &updatedBalance[i]
		}
		systemMetadata.InflationBalance = toStoreBalance

		context.Logger().
			Debug("Added balance for contract in inflation reward pool",
				"contract", contractAddress,
				"balanceAdded", inflationRewardByContractAddress[contractAddress],
				"totalBalance", updatedBalance)

		if err := gasTrackingKeeper.SetContractSystemMetadata(context, contractAddr, systemMetadata); err != nil {
			panic(err)
		}
	}
}

// distributeGasRewards distributes the calculated rewards to all the contracts owners.
func distributeGasRewards(context sdk.Context, rewardAddresses []string, rewardsByAddress map[string]sdk.DecCoins, gasTrackingKeeper gstKeeper.GasTrackingKeeper, rewardTransferKeeper RewardTransferKeeper) {
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

		err = rewardTransferKeeper.SendCoinsFromModuleToAccount(context, gstTypes.GasRewardCollector, rewardAddress, rewardsToBePayed)
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

		context.Logger().Debug("Reward allocation details:", "rewardPayed", rewardsToBePayed, "leftOverEntry", leftOverEntry.ContractRewards)
	}
}

type contractRewardData struct {
	totalGasRewards          sdk.DecCoins
	rewardAddresses          []string
	gasRewardByRewardAddress map[string]sdk.DecCoins

	totalInflationRewards            sdk.DecCoins
	contractAddresses                []string
	inflationRewardByContractAddress map[string]sdk.DecCoins
}

// getContractRewards returns the total rewards and the rewards per contract based on the calculations.
func getContractRewards(context sdk.Context, blockGasTracking gstTypes.BlockGasTracking, gasTrackingKeeper gstKeeper.GasTrackingKeeper, contractTotalInflationRewards sdk.DecCoin) contractRewardData {
	// To enforce a map iteration order. This isn't strictly necessary but is only
	// done to make this code more deterministic.
	rewardAddresses := make([]string, 0)
	contractAddresses := make([]string, 0)

	gasRewardByRewardAddress := make(map[string]sdk.DecCoins)
	inflationRewardByContractAddress := make(map[string]sdk.DecCoins)

	totalGasRewardsPerBlock := make(sdk.DecCoins, 0)
	totalInflationRewardsPerBlock := make(sdk.DecCoins, 0)
	for _, txTrackingInfo := range blockGasTracking.TxTrackingInfos {
		if !txTrackingInfo.IsEligibleForRewards {
			continue
		}

		// We generate empty coins based on the fees coins.
		totalGasRewardsInTx := make(sdk.DecCoins, 0)
		totalInflationRewardsInTx := make(sdk.DecCoins, 0)

		for _, contractTrackingInfo := range txTrackingInfo.ContractTrackingInfos {
			contractAddress, err := sdk.AccAddressFromBech32(contractTrackingInfo.Address)
			if err != nil {
				panic(err)
			}

			metadata, err := gasTrackingKeeper.GetContractMetadata(context, contractAddress)
			if err != nil {
				panic(err)
			}
			context.Logger().Debug("Got the metadata for contract", "contract", contractAddress, "metadata", metadata)

			gasRewards := make(sdk.DecCoins, 0, 0)
			inflationRewards := make(sdk.DecCoins, 0, 0)

			gasConsumedInContract := sdk.NewDecFromBigInt(gstTypes.ConvertUint64ToBigInt(contractTrackingInfo.OriginalSdkGas + contractTrackingInfo.OriginalVmGas))

			if gasTrackingKeeper.IsDappInflationRewardsEnabled(context) && context.BlockGasMeter().Limit() > 0 {
				blockGasLimit := sdk.NewDecFromBigInt(gstTypes.ConvertUint64ToBigInt(context.BlockGasMeter().Limit()))
				contractInflationReward := sdk.NewDecCoinFromDec(contractTotalInflationRewards.Denom, contractTotalInflationRewards.Amount.Mul(gasConsumedInContract).Quo(blockGasLimit))
				context.Logger().Debug("Calculated contract inflation rewards:", "contractAddress", contractAddress, "contractInflationReward", contractInflationReward)
				inflationRewards = inflationRewards.Add(contractInflationReward)
			}

			if !gasTrackingKeeper.IsGasRebateToUserEnabled(context) || !metadata.GasRebateToUser {
				maxGasAllowedInTx := sdk.NewDecFromBigInt(gstTypes.ConvertUint64ToBigInt(txTrackingInfo.MaxGasAllowed))

				// Calc premium fees
				var gasUsageForUsageRewards = gasConsumedInContract
				if metadata.CollectPremium && gasTrackingKeeper.IsContractPremiumEnabled(context) {
					premiumGas := gasConsumedInContract.
						Mul(sdk.NewDecFromBigInt(gstTypes.ConvertUint64ToBigInt(metadata.PremiumPercentageCharged))).
						QuoInt64(100)
					gasUsageForUsageRewards = gasUsageForUsageRewards.Add(premiumGas)
				}

				if gasTrackingKeeper.IsGasRebateToContractEnabled(context) {
					gasRebateRewards := make(sdk.DecCoins, 0)
					for _, rewardCoin := range txTrackingInfo.MaxContractRewards {
						gasRebateRewards = gasRebateRewards.Add(sdk.NewDecCoinFromDec(
							rewardCoin.Denom, rewardCoin.Amount.Mul(gasUsageForUsageRewards).Quo(maxGasAllowedInTx)))
					}
					context.Logger().
						Debug("Calculated contract gas rebate rewards:",
							"contractAddress", contractAddress, "contractGasReward", gasRebateRewards)
					gasRewards = gasRewards.Add(gasRebateRewards...)
				}
			} else {
				context.Logger().Debug("Contract is not eligible for gas rewards, skipping calculation.", "contractAddress", contractAddress)
			}

			if _, ok := inflationRewardByContractAddress[contractAddress.String()]; !ok {
				contractAddresses = append(contractAddresses, contractAddress.String())
				inflationRewardByContractAddress[contractAddress.String()] = inflationRewards
			} else {
				inflationRewardByContractAddress[contractAddress.String()] = inflationRewardByContractAddress[contractAddress.String()].Add(inflationRewards...)
			}

			if _, ok := gasRewardByRewardAddress[metadata.RewardAddress]; !ok {
				rewardAddresses = append(rewardAddresses, metadata.RewardAddress)
				gasRewardByRewardAddress[metadata.RewardAddress] = gasRewards
			} else {
				gasRewardByRewardAddress[metadata.RewardAddress] = gasRewardByRewardAddress[metadata.RewardAddress].Add(gasRewards...)
			}

			totalGasRewardsInTx = totalGasRewardsInTx.Add(gasRewards...)
			totalInflationRewardsInTx = totalInflationRewardsInTx.Add(inflationRewards...)

			if err = EmitContractRewardCalculationEvent(context, contractAddress.String(), gasConsumedInContract, inflationRewards, gasRewards, &metadata); err != nil {
				panic(err)
			}

			context.Logger().Debug("Calculated Contract rewards:", "contractAddress", contractAddress, "gasRewards", gasRewards, "inflationRewards", inflationRewards)
		}

		totalGasRewardsPerBlock = totalGasRewardsPerBlock.Add(totalGasRewardsInTx...)
		totalInflationRewardsPerBlock = totalInflationRewardsPerBlock.Add(totalInflationRewardsInTx...)
	}

	return contractRewardData{
		gasRewardByRewardAddress:         gasRewardByRewardAddress,
		inflationRewardByContractAddress: inflationRewardByContractAddress,
		totalGasRewards:                  totalGasRewardsPerBlock,
		totalInflationRewards:            totalInflationRewardsPerBlock,
		rewardAddresses:                  rewardAddresses,
		contractAddresses:                contractAddresses,
	}
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
