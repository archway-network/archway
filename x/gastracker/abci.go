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

func BeginBlock(context sdk.Context, _ abci.RequestBeginBlock, keeper GasTrackingKeeper, rewardTransferKeeper RewardTransferKeeper, mintParamsKeeper MintParamsKeeper) {
	blockTxDetails, err := keeper.GetCurrentBlockTrackingInfo(context)
	if err != nil {
		panic(err)
	}

	if err := keeper.TrackNewBlock(context, gstTypes.BlockGasTracking{}); err != nil {
		panic(err)
	}

	var calculatedGasConsumedInLastBlock uint64 = 0
	for _, txTrackingInfo := range blockTxDetails.TxTrackingInfos {
		for _, contractTrackingInfo := range txTrackingInfo.ContractTrackingInfos {
			calculatedGasConsumedInLastBlock += contractTrackingInfo.GasConsumed
		}
	}
	var totalGasConsumedInLastBlock = sdk.NewDecFromBigInt(ConvertUint64ToBigInt(calculatedGasConsumedInLastBlock))

	context.Logger().Info("Got the tracking for block", "BlockTxDetails", blockTxDetails)

	rewardsByAddress := make(map[string]sdk.DecCoins)

	totalContractRewardsPerBlock := make(sdk.DecCoins, 0)

	minter := mintParamsKeeper.GetMinter(context)
	params := mintParamsKeeper.GetParams(context)
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
			decGasUsage := sdk.NewDecFromBigInt(ConvertUint64ToBigInt(contractTrackingInfo.GasConsumed))

			contractRewards := make(sdk.DecCoins, len(txTrackingInfo.MaxContractRewards))
			for i, rewardCoin := range txTrackingInfo.MaxContractRewards {
				contractRewards[i] = sdk.NewDecCoinFromDec(rewardCoin.Denom, rewardCoin.Amount.Mul(decGasUsage).Quo(decGasLimit))
			}
			contractInflationReward := sdk.NewDecCoinFromDec(contractTotalInflationRewards.Denom, contractTotalInflationRewards.Amount.Mul(decGasUsage).Quo(totalGasConsumedInLastBlock))
			context.Logger().Info("Calculated contract inflation rewards:", "contractAddress", contractTrackingInfo.Address, "contractInflationReward", contractInflationReward)
			contractRewards = contractRewards.Add(contractInflationReward)

			if _, ok := rewardsByAddress[metadata.RewardAddress]; !ok {
				rewardsByAddress[metadata.RewardAddress] = contractRewards
			} else {
				rewardsByAddress[metadata.RewardAddress] = rewardsByAddress[metadata.RewardAddress].Add(contractRewards...)
			}
			totalContractRewardsInTx = totalContractRewardsInTx.Add(contractRewards...)

			context.Logger().Info("Calculated Contract rewards:", "contractAddress", contractTrackingInfo.Address, "contractRewards", contractRewards)
		}

		totalContractRewardsPerBlock = totalContractRewardsPerBlock.Add(totalContractRewardsInTx...)
	}

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

		err = rewardTransferKeeper.SendCoinsFromModuleToAccount(context, ContractRewardCollector, transferAddr, rewardsToBePayed)
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
}