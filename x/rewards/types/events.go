package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EmitContractMetadataSetEvent(ctx sdk.Context, contractAddr sdk.AccAddress, metadata ContractMetadata) {
	err := ctx.EventManager().EmitTypedEvent(&ContractMetadataSetEvent{
		ContractAddress: contractAddr.String(),
		Metadata:        metadata,
	})
	if err != nil {
		panic(fmt.Errorf("sending ContractMetadataSetEvent event: %w", err))
	}
}

func EmitContractRewardCalculationEvent(ctx sdk.Context, contractAddr sdk.AccAddress, gasConsumed uint64, inflationRewards sdk.Coin, feeRebateRewards sdk.Coins, metadata *ContractMetadata) {
	err := ctx.EventManager().EmitTypedEvent(&ContractRewardCalculationEvent{
		ContractAddress:  contractAddr.String(),
		GasConsumed:      gasConsumed,
		InflationRewards: inflationRewards,
		FeeRebateRewards: feeRebateRewards,
		Metadata:         metadata,
	})
	if err != nil {
		panic(fmt.Errorf("sending ContractRewardCalculationEvent event: %w", err))
	}
}

func EmitContractRewardDistributionEvent(ctx sdk.Context, contractAddr, rewardAddress sdk.AccAddress, rewards sdk.Coins) {
	err := ctx.EventManager().EmitTypedEvent(&ContractRewardDistributionEvent{
		ContractAddress: contractAddr.String(),
		RewardAddress:   rewardAddress.String(),
		Rewards:         rewards,
	})
	if err != nil {
		panic(fmt.Errorf("sending ContractRewardDistributionEvent event: %w", err))
	}
}

func EmitMinConsensusFeeSetEvent(ctx sdk.Context, fee sdk.DecCoin) {
	err := ctx.EventManager().EmitTypedEvent(&MinConsensusFeeSetEvent{
		Fee: fee,
	})
	if err != nil {
		panic(fmt.Errorf("sending MinConsensusFeeSetEvent event: %w", err))
	}
}
