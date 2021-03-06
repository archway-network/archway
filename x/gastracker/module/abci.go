package module

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	abci "github.com/tendermint/tendermint/abci/types"

	gstTypes "github.com/archway-network/archway/x/gastracker"
	"github.com/archway-network/archway/x/gastracker/keeper"
)

type RewardTransferKeeper interface {
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

type MintParamsKeeper interface {
	GetParams(ctx sdk.Context) (params mintTypes.Params)
	GetMinter(ctx sdk.Context) (minter mintTypes.Minter)
}

func BeginBlock(context sdk.Context, _ abci.RequestBeginBlock, gasTrackingKeeper keeper.Keeper, rewardTransferKeeper RewardTransferKeeper) {
	defer telemetry.ModuleMeasureSince(gstTypes.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	lastBlockGasTracking := resetBlockGasTracking(context, gasTrackingKeeper)

	params := gasTrackingKeeper.GetParams(context)

	if !params.GasTrackingSwitch { // No rewards or calculations should take place
		return
	}
	context.Logger().Debug("Got the tracking for block", "BlockTxDetails", lastBlockGasTracking)

	contractInflationaryRewards, err := gasTrackingKeeper.GetCurrentBlockDappInflationaryRewards(context)
	if err != nil {
		panic(err)
	}

	totalContractRewardsPerBlock, rewardAddresses, rewardsByAddress := getContractRewards(context, params, lastBlockGasTracking, gasTrackingKeeper, contractInflationaryRewards)

	// We need to commit pending metadata before we return but after we calculated rewards.
	commitPendingMetadata(context, gasTrackingKeeper)

	// Either the tx did not collect any fee or no contracts were executed
	// So, no need to continue execution
	if totalContractRewardsPerBlock == nil || totalContractRewardsPerBlock.IsZero() {
		return
	}

	distributeRewards(context, rewardAddresses, rewardsByAddress, gasTrackingKeeper, rewardTransferKeeper)
}

func commitPendingMetadata(context sdk.Context, gasTrackingKeeper keeper.Keeper) {
	numberOfEntriesCommitted, err := gasTrackingKeeper.CommitPendingContractMetadata(context)
	if err != nil {
		panic(err)
	}
	context.Logger().Debug("Committed pending metadata change", "NumberOfMetadataCommitted", numberOfEntriesCommitted)
}

// resetBlockGasTracking resets the current status and returns the last blockGasTracking
func resetBlockGasTracking(context sdk.Context, gasTrackingKeeper keeper.Keeper) gstTypes.BlockGasTracking {
	lastBlockGasTracking := gasTrackingKeeper.GetCurrentBlockTracking(context)

	gasTrackingKeeper.TrackNewBlock(context)

	return lastBlockGasTracking
}

// distributeRewards distributes the calculated rewards to all the contracts owners.
func distributeRewards(context sdk.Context, rewardAddresses []string, rewardsByAddress map[string]sdk.DecCoins, gasTrackingKeeper keeper.Keeper, rewardTransferKeeper RewardTransferKeeper) {
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

		err = rewardTransferKeeper.SendCoinsFromModuleToAccount(context, gstTypes.ContractRewardCollector, rewardAddress, rewardsToBePayed)
		if err != nil {
			panic(err)
		}

		leftOverEntry, err := gasTrackingKeeper.GetLeftOverRewardEntry(context, rewardAddress)
		if err != nil {
			panic(err)
		}

		gstTypes.EmitRewardPayingEvent(context, rewardAddressStr, rewardsToBePayed, leftOverEntry.ContractRewards)

		context.Logger().Debug("Reward allocation details:", "rewardPayed", rewardsToBePayed, "leftOverEntry", leftOverEntry.ContractRewards)
	}
}

// getContractRewards returns the total rewards and the rewards per contract based on the calculations.
func getContractRewards(context sdk.Context, params gstTypes.Params, blockGasTracking gstTypes.BlockGasTracking, gasTrackingKeeper keeper.Keeper, contractTotalInflationRewards sdk.DecCoin) (sdk.DecCoins, []string, map[string]sdk.DecCoins) {
	// To enforce a map iteration order. This isn't strictly necessary but is only
	// done to make this code more deterministic.
	rewardAddresses := make([]string, 0)
	rewardsByAddress := make(map[string]sdk.DecCoins)

	totalContractRewardsPerBlock := make(sdk.DecCoins, 0)
	for _, txTrackingInfo := range blockGasTracking.TxTrackingInfos {
		// We generate empty coins based on the fees coins.
		totalContractRewardsInTx := make(sdk.DecCoins, len(txTrackingInfo.MaxContractRewards))
		for i := range totalContractRewardsInTx {
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
			context.Logger().Debug("Got the metadata for contract", "contract", contractAddress, "metadata", metadata)

			contractRewards := make(sdk.DecCoins, 0, 0)

			gasConsumedInContract := sdk.NewDecFromBigInt(gstTypes.ConvertUint64ToBigInt(contractTrackingInfo.OriginalSdkGas + contractTrackingInfo.OriginalVmGas))

			if !params.DappInflationRewardsRatio.IsZero() && context.BlockGasMeter().Limit() > 0 {
				blockGasLimit := sdk.NewDecFromBigInt(gstTypes.ConvertUint64ToBigInt(context.BlockGasMeter().Limit()))
				contractInflationReward = sdk.NewDecCoinFromDec(contractTotalInflationRewards.Denom, contractTotalInflationRewards.Amount.Mul(gasConsumedInContract).Quo(blockGasLimit))
				context.Logger().Debug("Calculated contract inflation rewards:", "contractAddress", contractAddress, "contractInflationReward", contractInflationReward)
				contractRewards = contractRewards.Add(contractInflationReward)
			}

			if !params.GasRebateToUserSwitch || !metadata.GasRebateToUser {
				maxGasAllowedInTx := sdk.NewDecFromBigInt(gstTypes.ConvertUint64ToBigInt(txTrackingInfo.MaxGasAllowed))

				// Calc premium fees
				var gasUsageForUsageRewards = gasConsumedInContract
				if metadata.CollectPremium && params.ContractPremiumSwitch {
					premiumGas := gasConsumedInContract.
						Mul(sdk.NewDecFromBigInt(gstTypes.ConvertUint64ToBigInt(metadata.PremiumPercentageCharged))).
						QuoInt64(100)
					gasUsageForUsageRewards = gasUsageForUsageRewards.Add(premiumGas)
				}

				if !params.DappTxFeeRebateRatio.IsZero() {
					gasRebateRewards := make(sdk.DecCoins, 0)
					for _, rewardCoin := range txTrackingInfo.MaxContractRewards {
						gasRebateRewards = gasRebateRewards.Add(sdk.NewDecCoinFromDec(
							rewardCoin.Denom, rewardCoin.Amount.Mul(gasUsageForUsageRewards).Quo(maxGasAllowedInTx)))
					}
					context.Logger().
						Debug("Calculated contract gas rebate rewards:",
							"contractAddress", contractAddress, "contractGasReward", gasRebateRewards)
					contractRewards = contractRewards.Add(gasRebateRewards...)
				}
			} else {
				context.Logger().Debug("Contract is not eligible for gas rewards, skipping calculation.", "contractAddress", contractAddress)
			}

			if _, ok := rewardsByAddress[metadata.RewardAddress]; !ok {
				rewardAddresses = append(rewardAddresses, metadata.RewardAddress)
				rewardsByAddress[metadata.RewardAddress] = contractRewards
			} else {
				rewardsByAddress[metadata.RewardAddress] = rewardsByAddress[metadata.RewardAddress].Add(contractRewards...)
			}

			totalContractRewardsInTx = totalContractRewardsInTx.Add(contractRewards...)

			gstTypes.EmitContractRewardCalculationEvent(context, contractAddress.String(), gasConsumedInContract, contractInflationReward, contractRewards, metadata)

			context.Logger().Debug("Calculated Contract rewards:", "contractAddress", contractAddress, "contractRewards", contractRewards)
		}

		totalContractRewardsPerBlock = totalContractRewardsPerBlock.Add(totalContractRewardsInTx...)
	}

	return totalContractRewardsPerBlock, rewardAddresses, rewardsByAddress
}
