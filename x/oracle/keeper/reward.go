package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/oracle/types"
)

func (k Keeper) AllocateRewards(ctx sdk.Context, funderModule string, totalCoins sdk.Coins, votePeriods uint64) error {
	votePeriodCoins := make(sdk.Coins, len(totalCoins))
	for i, coin := range totalCoins {
		newCoin := sdk.NewCoin(coin.Denom, coin.Amount.QuoRaw(int64(votePeriods)))
		votePeriodCoins[i] = newCoin
	}

	id, err := k.RewardsID.Next(ctx)
	if err != nil {
		return err
	}
	k.Rewards.Set(ctx, id, types.Rewards{
		Id:          id,
		VotePeriods: votePeriods,
		Coins:       votePeriodCoins,
	})

	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, funderModule, types.ModuleName, totalCoins)
}

// rewardWinners gives out a portion of spread fees collected in the
// oracle reward pool to the oracle voters that voted faithfully.
func (k Keeper) rewardWinners(
	ctx sdk.Context,
	validatorPerformances types.ValidatorPerformances,
) {
	totalRewardWeight := validatorPerformances.TotalRewardWeight()
	if totalRewardWeight == 0 {
		return
	}

	var totalRewards sdk.DecCoins
	rewards := k.GatherRewardsForVotePeriod(ctx)
	totalRewards = totalRewards.Add(sdk.NewDecCoinsFromCoins(rewards...)...)

	var distributedRewards sdk.Coins
	for _, validatorPerformance := range validatorPerformances {
		validator, err := k.StakingKeeper.Validator(ctx, validatorPerformance.ValAddress)
		if err != nil {
			continue
		}

		rewardPortion, _ := totalRewards.MulDec(math.LegacyNewDec(validatorPerformance.RewardWeight).QuoInt64(totalRewardWeight)).TruncateDecimal()
		k.distrKeeper.AllocateTokensToValidator(ctx, validator, sdk.NewDecCoinsFromCoins(rewardPortion...))
		distributedRewards = distributedRewards.Add(rewardPortion...)
	}

	// Move distributed reward to distribution module
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.distrModuleName, distributedRewards)
	if err != nil {
		k.Logger(ctx).Error("Failed to send coins to distribution module", "err", err)
	}
}

// GatherRewardsForVotePeriod retrieves the pair rewards for the provided pair and current vote period.
func (k Keeper) GatherRewardsForVotePeriod(ctx sdk.Context) sdk.Coins {
	coins := sdk.NewCoins()
	// iterate over
	iter, err := k.Rewards.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	keys, err := iter.Keys()
	if err != nil {
		panic(err)
	}
	for _, rewardId := range keys {
		pairReward, err := k.Rewards.Get(ctx, rewardId)
		if err != nil {
			k.Logger(ctx).Error("Failed to get reward", "err", err)
			continue
		}
		coins = coins.Add(pairReward.Coins...)

		// Decrease the remaining vote periods of the PairReward.
		pairReward.VotePeriods -= 1
		if pairReward.VotePeriods == 0 {
			// If the distribution period count drops to 0: the reward instance is removed.
			err := k.Rewards.Remove(ctx, rewardId)
			if err != nil {
				k.Logger(ctx).Error("Failed to delete pair reward", "err", err)
			}
		} else {
			k.Rewards.Set(ctx, rewardId, pairReward)
		}
	}

	return coins
}
