package gastracker

import "github.com/cosmos/cosmos-sdk/types"

func EmitRewardPayingEvent(ctx types.Context, rewardAddress string, rewardsPayed types.Coins, leftOverRewards []*types.DecCoin) {
	rewards := make([]*types.Coin, len(rewardsPayed))
	for i := range rewards {
		rewards[i] = &rewardsPayed[i]
	}

	err := ctx.EventManager().EmitTypedEvent(&RewardDistributionEvent{
		RewardAddress:   rewardAddress,
		ContractRewards: rewards,
		LeftoverRewards: leftOverRewards,
	})
	if err != nil {
		panic(err)
	}
}

func EmitContractRewardCalculationEvent(context types.Context, contractAddress string, gasConsumed types.Dec, inflationReward types.DecCoin, contractRewards types.DecCoins, metadata *ContractInstanceMetadata) {
	rewards := make([]*types.DecCoin, len(contractRewards))
	for i := range rewards {
		rewards[i] = &contractRewards[i]
	}

	err := context.EventManager().EmitTypedEvent(&ContractRewardCalculationEvent{
		ContractAddress:  contractAddress,
		GasConsumed:      gasConsumed.RoundInt().Uint64(),
		InflationRewards: &inflationReward,
		ContractRewards:  rewards,
		Metadata:         metadata,
	})
	if err != nil {
		panic(err)
	}
}
