package gastracker

import "github.com/cosmos/cosmos-sdk/types"

func EmitRewardPayingEvent(ctx types.Context, rewardAddress string, rewardsPayed types.Coins, leftOverRewards types.DecCoins) {
	err := ctx.EventManager().EmitTypedEvent(&RewardDistributionEvent{
		RewardAddress:   rewardAddress,
		ContractRewards: rewardsPayed,
		LeftoverRewards: leftOverRewards,
	})
	if err != nil {
		panic(err)
	}
}

func EmitContractRewardCalculationEvent(context types.Context, contractAddress string, gasConsumed types.Dec, inflationReward types.DecCoin, contractRewards types.DecCoins, metadata ContractInstanceMetadata) {
	err := context.EventManager().EmitTypedEvent(&ContractRewardCalculationEvent{
		ContractAddress:  contractAddress,
		GasConsumed:      gasConsumed.RoundInt().Uint64(),
		InflationRewards: inflationReward,
		ContractRewards:  contractRewards,
		Metadata:         metadata,
	})
	if err != nil {
		panic(err)
	}
}
