package gastracker

import (
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	abci "github.com/tendermint/tendermint/abci/types"
)

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

func BeginBlock(context sdk.Context, block abci.RequestBeginBlock, keeper GasTrackingKeeper, bankKeeper bankkeeper.Keeper, mintKeeper mintkeeper.Keeper) {
	blockTxDetails, err := keeper.GetCurrentBlockTrackingInfo(context)
	if err != nil {
		panic(err)
	}

	context.Logger().Info("Got the tracking for block", "BlockTxDetails", blockTxDetails)

	rewardsByAddress := make(map[string]sdk.DecCoins)

	totalContractRewardsPerBlock := make(sdk.DecCoins, 0)

	minter := mintKeeper.GetMinter(context)
	params := mintKeeper.GetParams(context)
	totalInflationFee := sdk.NewDecCoinFromCoin(minter.BlockProvision(params))
	// TODO: Take the percentage value from governance
	contractTotalInflationRewards := sdk.NewDecCoinFromDec(totalInflationFee.Denom, totalInflationFee.Amount.MulInt64(20).QuoInt64(100))

	for _, txTrackingInfo := range blockTxDetails.TxTrackingInfos {
		totalContractRewardsInTx := make(sdk.DecCoins, len(txTrackingInfo.MaxContractRewards))
		for i, _ := range totalContractRewardsInTx {
			totalContractRewardsInTx[i] = sdk.NewDecCoin(txTrackingInfo.MaxContractRewards[i].Denom, sdk.NewInt(0))
		}

		for _, contractTrackingInfo := range txTrackingInfo.ContractTrackingInfos {
			if !contractTrackingInfo.IsEligibleForReward {
				context.Logger().Info("Contract is not eligible for reward, skipping calculation.", "contractAddress", contractTrackingInfo.Address)
				continue
			}

			metadata, err := keeper.GetNewContractMetadata(context, contractTrackingInfo.Address)
			if err != nil {
				panic(err)
			}
			context.Logger().Info("Got the metadata", "Metadata", metadata)

			decGasLimit := sdk.NewDecFromBigInt(ConvertUint64ToBigInt(txTrackingInfo.MaxGasAllowed))
			gasUsageForInflationRewards := sdk.NewDecFromBigInt(ConvertUint64ToBigInt(contractTrackingInfo.GasConsumed))

			var gasUsageForUsageRewards sdk.Dec
			if metadata.CollectPremium {
				premiumGas := gasUsageForInflationRewards.Mul(sdk.NewDecFromBigInt(ConvertUint64ToBigInt(metadata.PremiumPercentageCharged))).QuoInt64(100)
				gasUsageForUsageRewards = gasUsageForInflationRewards.Add(premiumGas)
			} else {
				gasUsageForUsageRewards = gasUsageForInflationRewards
			}

			context.Logger().Info("Gas usage for reward calculation:", "gasUsageForUsageRewards", gasUsageForUsageRewards, "gasUsageForInflationRewards", gasUsageForInflationRewards)

			contractRewards := make(sdk.DecCoins, len(txTrackingInfo.MaxContractRewards))
			for i, rewardCoin := range txTrackingInfo.MaxContractRewards {
				contractRewards[i] = sdk.NewDecCoinFromDec(rewardCoin.Denom, rewardCoin.Amount.Mul(gasUsageForUsageRewards).Quo(decGasLimit))
			}
			contractInflationReward := sdk.NewDecCoinFromDec(contractTotalInflationRewards.Denom, contractTotalInflationRewards.Amount.Mul(gasUsageForInflationRewards).Quo(decGasLimit))
			context.Logger().Info("Calculated contract inflation rewards:", "contractAddress", contractTrackingInfo.Address, "contractInflationReward", contractInflationReward)
			contractRewards = contractRewards.Add(contractInflationReward)

			if currentRewardData, ok := rewardsByAddress[metadata.RewardAddress]; !ok {
				rewardsByAddress[metadata.RewardAddress] = contractRewards
			} else {
				rewardsByAddress[metadata.RewardAddress] = rewardsByAddress[metadata.RewardAddress].Add(currentRewardData...)
			}
			totalContractRewardsInTx = totalContractRewardsInTx.Add(contractRewards...)

			context.Logger().Info("Calculated Contract rewards:", "contractAddress", contractTrackingInfo.Address, "contractRewards", contractRewards)
		}

		totalContractRewardsPerBlock = totalContractRewardsPerBlock.Add(totalContractRewardsInTx...)
	}

	totalFeeToBeCollected := make(sdk.Coins, len(totalContractRewardsPerBlock))
	for i := range totalFeeToBeCollected {
		totalFeeToBeCollected[i] = sdk.NewCoin(totalContractRewardsPerBlock[i].Denom, totalContractRewardsPerBlock[i].Amount.Ceil().RoundInt())
	}

	err = bankKeeper.SendCoinsFromModuleToModule(context, authTypes.FeeCollectorName, ContractRewardCollector, totalFeeToBeCollected)
	if err != nil {
		panic(err)
	}

	for rewardAddress, rewards := range rewardsByAddress {
		// TODO: We should take leftOverThreshold from governance
		rewardsToBePayed, err := keeper.CreateOrMergeLeftOverRewardEntry(context, rewardAddress, rewards, 1)
		if err != nil {
			panic(err)
		}

		transferAddr, err := sdk.AccAddressFromBech32(rewardAddress)
		if err != nil {
			panic(err)
		}

		err = bankKeeper.SendCoinsFromModuleToAccount(context, ContractRewardCollector, transferAddr, rewardsToBePayed)
		if err != nil {
			panic(err)
		}

		leftOverEntry, err := keeper.GetLeftOverRewardEntry(context, rewardAddress)
		if err != nil {
			panic(err)
		}


		err = EmitRewardPayingEvent(context, rewardAddress, rewardsToBePayed, leftOverEntry.ContractRewards)
		if err != nil {
			panic(err)
		}

		context.Logger().Info("Reward allocation details:", "rewardPayed", rewardsToBePayed, "leftOverEntry", leftOverEntry.ContractRewards)
	}

	var newBlockTxDetails gstTypes.BlockGasTracking
	if err := keeper.TrackNewBlock(context, newBlockTxDetails); err != nil {
		panic(err)
	}
}