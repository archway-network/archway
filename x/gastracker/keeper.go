package gastracker

import (
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ GasTrackingKeeper = &Keeper{}

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

type Keeper struct {
	key sdk.StoreKey
}

func (k *Keeper) CreateOrMergeLeftOverRewardEntry(ctx sdk.Context, rewardAddress string, contractRewards sdk.DecCoins, leftOverThreshold uint64)  (sdk.Coins, error) {
	contractRewards = contractRewards.Sort()

	gstKvStore := ctx.KVStore(k.key)

	var rewardEntry gstTypes.LeftOverRewardEntry
	rewardsToBeDistributed := make(sdk.Coins, 0)

	bz := gstKvStore.Get(gstTypes.GetRewardEntryKey(rewardAddress))
	if bz != nil {
		err := rewardEntry.Unmarshal(bz)
		if err != nil {
			return rewardsToBeDistributed, err
		}

		previousRewards := make(sdk.DecCoins, len(rewardEntry.ContractRewards))
		for i := range rewardEntry.ContractRewards {
			previousRewards[i] = *rewardEntry.ContractRewards[i]
		}

		updatedRewards := previousRewards.Add(contractRewards...)

		rewardEntry.ContractRewards = make([]*sdk.DecCoin, len(updatedRewards))
		for i := range rewardEntry.ContractRewards {
			rewardEntry.ContractRewards[i] = &updatedRewards[i]
		}

	} else {
		rewardEntry.ContractRewards = make([]*sdk.DecCoin, len(contractRewards))
		for i := range contractRewards {
			rewardEntry.ContractRewards[i] = &contractRewards[i]
		}
	}

	// Reallocate to length of rewardEntry.ContractRewards
	rewardsToBeDistributed = make(sdk.Coins, len(rewardEntry.ContractRewards))

	leftOverDec := sdk.NewDecFromBigInt(ConvertUint64ToBigInt(leftOverThreshold))
	rewardIndex := 0
	for i := range rewardEntry.ContractRewards {
		truncatedDec := rewardEntry.ContractRewards[i].Amount.TruncateDec()
		if truncatedDec.GTE(leftOverDec) {
			rewardsToBeDistributed[rewardIndex] = sdk.NewCoin(rewardEntry.ContractRewards[i].Denom, truncatedDec.TruncateInt())
			rewardIndex += 1
			rewardEntry.ContractRewards[i].Amount = rewardEntry.ContractRewards[i].Amount.Sub(truncatedDec)
		}
	}
	// Resize rewards slice to only elements that are initialized above
	rewardsToBeDistributed = rewardsToBeDistributed[:rewardIndex]

	bz, err := rewardEntry.Marshal()
	if err != nil {
		return rewardsToBeDistributed, err
	}

	gstKvStore.Set(gstTypes.GetRewardEntryKey(rewardAddress), bz)
	return rewardsToBeDistributed, nil
}

func (k *Keeper) GetLeftOverRewardEntry(ctx sdk.Context, rewardAddress string) (gstTypes.LeftOverRewardEntry, error) {
	gstKvStore := ctx.KVStore(k.key)

	var rewardEntry gstTypes.LeftOverRewardEntry

	bz := gstKvStore.Get(gstTypes.GetRewardEntryKey(rewardAddress))
	if bz == nil {
		return rewardEntry, gstTypes.ErrRewardEntryNotFound
	}

	err := rewardEntry.Unmarshal(bz)
	if err != nil {
		return rewardEntry, err
	}

	return rewardEntry, nil
}

func (k *Keeper) GetNewContractMetadata(ctx sdk.Context, address string) (gstTypes.ContractInstanceMetadata, error) {
	gstKvStore := ctx.KVStore(k.key)

	var contractInstanceMetadata gstTypes.ContractInstanceMetadata

	bz := gstKvStore.Get(gstTypes.GetContractInstanceMetadataKey(address))
	if bz == nil {
		return contractInstanceMetadata, gstTypes.ErrContractInstanceMetadataNotFound
	}

	err := contractInstanceMetadata.Unmarshal(bz)
	return contractInstanceMetadata, err
}

func (k *Keeper) AddNewContractMetadata(ctx sdk.Context, address string, metadata gstTypes.ContractInstanceMetadata) error {
	gstKvStore := ctx.KVStore(k.key)

	bz, err := metadata.Marshal()
	if err != nil {
		return err
	}
	gstKvStore.Set(gstTypes.GetContractInstanceMetadataKey(address), bz)
	return nil
}

func NewGasTrackingKeeper(key sdk.StoreKey) *Keeper {
	return &Keeper{key: key}
}

func (k *Keeper) TrackNewBlock(ctx sdk.Context, blockGasTracking gstTypes.BlockGasTracking) error {
	gstKvStore := ctx.KVStore(k.key)

	bz, err := blockGasTracking.Marshal()
	if err != nil {
		return err
	}
	gstKvStore.Set([]byte(gstTypes.CurrentBlockTrackingKey), bz)
	return nil
}

func (k *Keeper) GetCurrentBlockTrackingInfo(ctx sdk.Context) (gstTypes.BlockGasTracking, error) {
	gstKvStore := ctx.KVStore(k.key)

	var currentBlockTracking gstTypes.BlockGasTracking
	bz := gstKvStore.Get([]byte(gstTypes.CurrentBlockTrackingKey))
	if bz == nil {
		return currentBlockTracking, gstTypes.ErrBlockTrackingDataNotFound
	}
	err := currentBlockTracking.Unmarshal(bz)
	return currentBlockTracking, err
}

func (k *Keeper) TrackNewTx(ctx sdk.Context, fee []*sdk.DecCoin, gasLimit uint64)  error {
	gstKvStore := ctx.KVStore(k.key)

	var currentTxGasTracking gstTypes.TransactionTracking
	currentTxGasTracking.MaxContractRewards = fee
	currentTxGasTracking.MaxGasAllowed = gasLimit
	bz, err := currentTxGasTracking.Marshal()
	if err != nil {
		return err
	}

	bz = gstKvStore.Get([]byte(gstTypes.CurrentBlockTrackingKey))
	if bz == nil && ctx.BlockHeight() > 1 {
		return gstTypes.ErrBlockTrackingDataNotFound
	}
	var currentBlockTracking gstTypes.BlockGasTracking
	err = currentBlockTracking.Unmarshal(bz)
	if err != nil {
		return err
	}
	currentBlockTracking.TxTrackingInfos = append(currentBlockTracking.TxTrackingInfos, &currentTxGasTracking)
	bz, err = currentBlockTracking.Marshal()
	if err != nil {
		return err
	}
	gstKvStore.Set([]byte(gstTypes.CurrentBlockTrackingKey), bz)
	return nil
}

func (k *Keeper) TrackContractGasUsage(ctx sdk.Context, contractAddress string, gasUsed uint64, operation gstTypes.ContractOperation, isEligibleForReward bool) error {
	gstKvStore := ctx.KVStore(k.key)
	bz := gstKvStore.Get([]byte(gstTypes.CurrentBlockTrackingKey))
	if bz == nil {
		return gstTypes.ErrBlockTrackingDataNotFound
	}
	var currentBlockGasTracking gstTypes.BlockGasTracking
	err := currentBlockGasTracking.Unmarshal(bz)
	if err != nil {
		return err
	}

	txsLen := len(currentBlockGasTracking.TxTrackingInfos)
	if txsLen == 0 {
		return gstTypes.ErrTxTrackingDataNotFound
	}
	currentTxGasTracking := currentBlockGasTracking.TxTrackingInfos[txsLen - 1]
	currentBlockGasTracking.TxTrackingInfos[txsLen - 1].ContractTrackingInfos = append(currentTxGasTracking.ContractTrackingInfos, &gstTypes.ContractGasTracking{
		Address:     contractAddress,
		GasConsumed: gasUsed,
		Operation: operation,
		IsEligibleForReward: isEligibleForReward,
	})
	bz, err = currentBlockGasTracking.Marshal()
	if err != nil {
		return err
	}

	gstKvStore.Set([]byte(gstTypes.CurrentBlockTrackingKey), bz)
	return nil
}



