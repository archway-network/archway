# Keeper
The gastracker module provides an exproted keeper interface that can be passed on to other modules that want to:

- Read/Write contract metadata
- Track gas consumption of the block with granularity upto individual contracts


## Keeper
``go
type GasTrackingKeeper interface {
	TrackNewTx(ctx sdk.Context, fee []*sdk.DecCoin, gasLimit uint64)  error
	TrackContractGasUsage(ctx sdk.Context, contractAddress string, gasUsed uint64, operation gstTypes.ContractOperation, isEligibleForReward bool) error
	GetCurrentBlockTrackingInfo(ctx sdk.Context) (gstTypes.BlockGasTracking, error)
	TrackNewBlock(ctx sdk.Context, blockGasTracking gstTypes.BlockGasTracking) error
	AddNewContractMetadata(ctx sdk.Context, address string, metadata gstTypes.ContractInstanceMetadata) error
	GetNewContractMetadata(ctx sdk.Context, address string) (gstTypes.ContractInstanceMetadata, error)

	CreateOrMergeLeftOverRewardEntry(ctx sdk.Context, rewardAddress string, contractRewards sdk.DecCoins, leftOverThreshold uint64) (sdk.Coins, error)
	GetLeftOverRewardEntry(ctx sdk.Context, rewardAddress string) (gstTypes.LeftOverRewardEntry, error)
}
```

