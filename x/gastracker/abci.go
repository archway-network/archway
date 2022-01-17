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

func BeginBlock(context sdk.Context, _ abci.RequestBeginBlock, gasTrackingKeeper GasTrackingKeeper, rewardTransferKeeper RewardTransferKeeper, mintParamsKeeper MintParamsKeeper) {
	lastBlockGasTracking, err := updateBlockGasTracking(context, gasTrackingKeeper)

	if !gasTrackingKeeper.IsGasTrackingEnabled(context) { // No rewards or calculations should take place
		return
	}
	context.Logger().Info("Got the tracking for block", "BlockTxDetails", lastBlockGasTracking)

	contractTotalInflationRewards := getContractInflationRewardsPerBlock(context, mintParamsKeeper) // 20% of the rewards distributed on every block

	totalContractRewardsPerBlock, rewardAddresses, rewardsByAddress := getContractRewardsPerBlock(context, lastBlockGasTracking, gasTrackingKeeper, contractTotalInflationRewards)
	// Either the tx did not collect any fee or no contracts were executed
	// So, no need to continue execution
	if totalContractRewardsPerBlock == nil || totalContractRewardsPerBlock.IsZero() {
		return
	}

	totalFeeToBeCollected := make(sdk.Coins, len(totalContractRewardsPerBlock))
	for i := range totalFeeToBeCollected {
		totalFeeToBeCollected[i] = sdk.NewCoin(totalContractRewardsPerBlock[i].Denom, totalContractRewardsPerBlock[i].Amount.Ceil().RoundInt())
	}

	err = rewardTransferKeeper.SendCoinsFromModuleToModule(context, authTypes.FeeCollectorName, ContractRewardCollector, totalFeeToBeCollected)
	if err != nil {
		panic(err)
	}

	distributeRewards(context, rewardAddresses, rewardsByAddress, gasTrackingKeeper, rewardTransferKeeper)
}

// updateBlockGasTracking saves the current status and returns the last blockGasTracking
func updateBlockGasTracking(context sdk.Context, gasTrackingKeeper GasTrackingKeeper) (gstTypes.BlockGasTracking, error) {
	lastBlockGasTracking, err := getCurrentBlockGasTracking(context, gasTrackingKeeper)

	// TODO is tracking an empty block gas tracking mandatory?
	if err := gasTrackingKeeper.TrackNewBlock(context, gstTypes.BlockGasTracking{}); err != nil {
		panic(err)
	}
	return lastBlockGasTracking, err
}

// getCurrentBlockGasTracking returns the actual block gas tracking, panics if empty and block height is bigger than one.
func getCurrentBlockGasTracking(context sdk.Context, gasTrackingKeeper GasTrackingKeeper) (gstTypes.BlockGasTracking, error) {
	currentBlockTrackingInfo, err := gasTrackingKeeper.GetCurrentBlockGasTracking(context)
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

	return currentBlockTrackingInfo, err
}

// distributeRewards distributes the calculated rewards to all the contracts.
func distributeRewards(context sdk.Context, rewardAddresses []string, rewardsByAddress map[string]sdk.DecCoins, gasTrackingKeeper GasTrackingKeeper, rewardTransferKeeper RewardTransferKeeper) {
	for _, rewardAddress := range rewardAddresses {
		rewards := rewardsByAddress[rewardAddress]
		// TODO: We should take leftOverThreshold from governance
		rewardsToBePayed, err := gasTrackingKeeper.CreateOrMergeLeftOverRewardEntry(context, rewardAddress, rewards, 1)
		if err != nil {
			panic(err)
		}

		transferAddr, err := sdk.AccAddressFromBech32(rewardAddress)
		if err != nil {
			panic(err)
		}

		err = rewardTransferKeeper.SendCoinsFromModuleToAccount(context, ContractRewardCollector, transferAddr, rewardsToBePayed)
		if err != nil {
			panic(err)
		}

		leftOverEntry, err := gasTrackingKeeper.GetLeftOverRewardEntry(context, rewardAddress)
		if err != nil {
			panic(err)
		}

		err = EmitRewardPayingEvent(context, rewardAddress, rewardsToBePayed, leftOverEntry.ContractRewards)
		if err != nil {
			panic(err)
		}

		context.Logger().Info("Reward allocation details:", "rewardPayed", rewardsToBePayed, "leftOverEntry", leftOverEntry.ContractRewards)
	}
}

func getContractRewardsPerBlock(context sdk.Context, blockGasTracking gstTypes.BlockGasTracking, gasTrackingKeeper GasTrackingKeeper, contractTotalInflationRewards sdk.DecCoin) (sdk.DecCoins, []string, map[string]sdk.DecCoins) {
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
			if !contractTrackingInfo.IsEligibleForReward {
				context.Logger().Info("Contract is not eligible for reward, skipping calculation.", "contractAddress", contractTrackingInfo.Address)
				continue
			}

			decGasLimit := sdk.NewDecFromBigInt(ConvertUint64ToBigInt(txTrackingInfo.MaxGasAllowed))
			gasConsumedInContract := sdk.NewDecFromBigInt(ConvertUint64ToBigInt(contractTrackingInfo.GasConsumed))

			metadata, err := gasTrackingKeeper.GetNewContractMetadata(context, contractTrackingInfo.Address)
			if err != nil {
				panic(err)
			}
			context.Logger().Info("Got the metadata", "Metadata", metadata)

			// Calc premium fees
			var gasUsageForUsageRewards = gasConsumedInContract
			if metadata.CollectPremium && gasTrackingKeeper.IsContractPremiumEnabled(context) {
				premiumGas := gasConsumedInContract.
					Mul(sdk.NewDecFromBigInt(ConvertUint64ToBigInt(metadata.PremiumPercentageCharged))).
					QuoInt64(100)
				gasUsageForUsageRewards = gasUsageForUsageRewards.Add(premiumGas)
			}

			contractRewards := make(sdk.DecCoins, 0, len(txTrackingInfo.MaxContractRewards))
			if gasTrackingKeeper.IsGasRebateEnabled(context) {
				for _, rewardCoin := range txTrackingInfo.MaxContractRewards {
					contractRewards = append(contractRewards, sdk.NewDecCoinFromDec(
						rewardCoin.Denom, rewardCoin.Amount.Mul(gasUsageForUsageRewards).Quo(decGasLimit)))
				}
				context.Logger().
					Info("Calculated contract gas rebate rewards:",
						"contractAddress", contractTrackingInfo.Address, "contractGasReward", contractRewards)
			}

			if gasTrackingKeeper.IsDappInflationRewardsEnabled(context) {
				contractInflationReward := sdk.NewDecCoinFromDec(contractTotalInflationRewards.Denom, contractTotalInflationRewards.Amount.Mul(gasConsumedInContract).Quo(blockGasTracking.GetGasConsumed()))
				context.Logger().Info("Calculated contract inflation rewards:", "contractAddress", contractTrackingInfo.Address, "contractInflationReward", contractInflationReward)
				if !contractRewards.IsZero() {
					contractRewards = contractRewards.Add(contractInflationReward)
				} else {
					contractRewards = append(contractRewards, contractInflationReward)
				}
			}

			if _, ok := rewardsByAddress[metadata.RewardAddress]; !ok {
				rewardAddresses = append(rewardAddresses, metadata.RewardAddress)
				rewardsByAddress[metadata.RewardAddress] = contractRewards
			} else {
				rewardsByAddress[metadata.RewardAddress] = rewardsByAddress[metadata.RewardAddress].Add(contractRewards...)
			}
			totalContractRewardsInTx = totalContractRewardsInTx.Add(contractRewards...)

			context.Logger().Info("Calculated Contract rewards:", "contractAddress", contractTrackingInfo.Address, "contractRewards", contractRewards)
		}

		totalContractRewardsPerBlock = totalContractRewardsPerBlock.Add(totalContractRewardsInTx...)
	}

	return totalContractRewardsPerBlock, rewardAddresses, rewardsByAddress
}

// getContractInflationRewardsPerBlock returns the percentage of the block rewards that are dedicated to contracts
// TODO now is 20% of the block rewards hardcoded.
func getContractInflationRewardsPerBlock(context sdk.Context, mintParamsKeeper MintParamsKeeper) sdk.DecCoin {
	totalInflationRatePerBlock := getInflationRatePerBlock(context, mintParamsKeeper)

	// TODO: Take the percentage value from governance
	contractTotalInflationRewards := sdk.NewDecCoinFromDec(totalInflationRatePerBlock.Denom, totalInflationRatePerBlock.Amount.MulInt64(20).QuoInt64(100))

	return contractTotalInflationRewards
}

// getInflationRatePerBlock returns the inflation per block. (Annual Inflation / NumblocksPerYear)
func getInflationRatePerBlock(context sdk.Context, mintParamsKeeper MintParamsKeeper) sdk.DecCoin {
	minter := mintParamsKeeper.GetMinter(context)
	params := mintParamsKeeper.GetParams(context)
	totalInflationFee := sdk.NewDecCoinFromCoin(minter.BlockProvision(params))

	return totalInflationFee
}
