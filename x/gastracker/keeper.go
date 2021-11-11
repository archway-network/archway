package gastracker

import (
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsTypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ GasTrackingKeeper = &Keeper{}

type GasTrackingKeeper interface {
	TrackNewTx(ctx sdk.Context, fee []*sdk.DecCoin, gasLimit uint64) error
	TrackContractGasUsage(ctx sdk.Context, contractAddress string, gasUsed uint64, operation gstTypes.ContractOperation, isEligibleForReward bool) error
	GetCurrentBlockTrackingInfo(ctx sdk.Context) (gstTypes.BlockGasTracking, error)
	TrackNewBlock(ctx sdk.Context, blockGasTracking gstTypes.BlockGasTracking) error
	AddNewContractMetadata(ctx sdk.Context, address string, metadata gstTypes.ContractInstanceMetadata) error
	GetNewContractMetadata(ctx sdk.Context, address string) (gstTypes.ContractInstanceMetadata, error)

	CreateOrMergeLeftOverRewardEntry(ctx sdk.Context, rewardAddress string, contractRewards sdk.DecCoins, leftOverThreshold uint64) (sdk.Coins, error)
	GetLeftOverRewardEntry(ctx sdk.Context, rewardAddress string) (gstTypes.LeftOverRewardEntry, error)
}

type Keeper struct {
	key        sdk.StoreKey
	appCodec   codec.Marshaler
	ParamStore paramsTypes.Subspace
}

func (k *Keeper) CreateOrMergeLeftOverRewardEntry(ctx sdk.Context, rewardAddress string, contractRewards sdk.DecCoins, leftOverThreshold uint64) (sdk.Coins, error) {
	contractRewards = contractRewards.Sort()

	gstKvStore := ctx.KVStore(k.key)

	var rewardEntry gstTypes.LeftOverRewardEntry
	var updatedRewards sdk.DecCoins
	var rewardsToBeDistributed sdk.Coins

	bz := gstKvStore.Get(gstTypes.GetRewardEntryKey(rewardAddress))
	if bz != nil {
		err := k.appCodec.UnmarshalBinaryBare(bz, &rewardEntry)
		if err != nil {
			return rewardsToBeDistributed, err
		}
		previousRewards := make(sdk.DecCoins, len(rewardEntry.ContractRewards))
		for i := range previousRewards {
			previousRewards[i] = *rewardEntry.ContractRewards[i]
		}
		updatedRewards = previousRewards.Add(contractRewards...)
	} else {
		updatedRewards = contractRewards
	}

	rewardsToBeDistributed = make(sdk.Coins, len(updatedRewards))
	distributionRewardIndex := 0

	leftOverContractRewards := make(sdk.DecCoins, len(updatedRewards))
	leftOverRewardIndex := 0

	leftOverDec := sdk.NewDecFromBigInt(ConvertUint64ToBigInt(leftOverThreshold))

	for i := range updatedRewards {
		if updatedRewards[i].Amount.GTE(leftOverDec) {
			distributionAmount := updatedRewards[i].Amount.TruncateInt()
			leftOverAmount := updatedRewards[i].Amount.Sub(distributionAmount.ToDec())
			if !leftOverAmount.IsZero() {
				leftOverContractRewards[leftOverRewardIndex] = sdk.NewDecCoinFromDec(updatedRewards[i].Denom, leftOverAmount)
				leftOverRewardIndex += 1
			}
			rewardsToBeDistributed[distributionRewardIndex] = sdk.NewCoin(updatedRewards[i].Denom, distributionAmount)
			distributionRewardIndex += 1
		} else {
			leftOverContractRewards[leftOverRewardIndex] = updatedRewards[i]
			leftOverRewardIndex += 1
		}
	}

	rewardsToBeDistributed = rewardsToBeDistributed[:distributionRewardIndex]
	leftOverContractRewards = leftOverContractRewards[:leftOverRewardIndex]

	rewardEntry.ContractRewards = make([]*sdk.DecCoin, len(leftOverContractRewards))
	for i := range leftOverContractRewards {
		rewardEntry.ContractRewards[i] = &leftOverContractRewards[i]
	}

	bz, err := k.appCodec.MarshalBinaryBare(&rewardEntry)
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

	err := k.appCodec.UnmarshalBinaryBare(bz, &rewardEntry)
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

	err := k.appCodec.UnmarshalBinaryBare(bz, &contractInstanceMetadata)
	return contractInstanceMetadata, err
}

func (k *Keeper) AddNewContractMetadata(ctx sdk.Context, address string, metadata gstTypes.ContractInstanceMetadata) error {
	gstKvStore := ctx.KVStore(k.key)

	bz, err := k.appCodec.MarshalBinaryBare(&metadata)
	if err != nil {
		return err
	}
	gstKvStore.Set(gstTypes.GetContractInstanceMetadataKey(address), bz)
	return nil
}

func NewGasTrackingKeeper(key sdk.StoreKey, appCodec codec.Marshaler) *Keeper {
	return &Keeper{key: key, appCodec: appCodec}
}

func (k *Keeper) TrackNewBlock(ctx sdk.Context, blockGasTracking gstTypes.BlockGasTracking) error {
	gstKvStore := ctx.KVStore(k.key)

	bz, err := k.appCodec.MarshalBinaryBare(&blockGasTracking)
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
	err := k.appCodec.UnmarshalBinaryBare(bz, &currentBlockTracking)
	return currentBlockTracking, err
}

func (k *Keeper) TrackNewTx(ctx sdk.Context, fee []*sdk.DecCoin, gasLimit uint64) error {
	gstKvStore := ctx.KVStore(k.key)

	var currentTxGasTracking gstTypes.TransactionTracking
	currentTxGasTracking.MaxContractRewards = fee
	currentTxGasTracking.MaxGasAllowed = gasLimit
	bz, err := k.appCodec.MarshalBinaryBare(&currentTxGasTracking)
	if err != nil {
		return err
	}

	bz = gstKvStore.Get([]byte(gstTypes.CurrentBlockTrackingKey))
	if bz == nil && ctx.BlockHeight() > 1 {
		return gstTypes.ErrBlockTrackingDataNotFound
	}
	var currentBlockTracking gstTypes.BlockGasTracking
	err = k.appCodec.UnmarshalBinaryBare(bz, &currentBlockTracking)
	if err != nil {
		return err
	}
	currentBlockTracking.TxTrackingInfos = append(currentBlockTracking.TxTrackingInfos, &currentTxGasTracking)
	bz, err = k.appCodec.MarshalBinaryBare(&currentBlockTracking)
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
	err := k.appCodec.UnmarshalBinaryBare(bz, &currentBlockGasTracking)
	if err != nil {
		return err
	}

	txsLen := len(currentBlockGasTracking.TxTrackingInfos)
	if txsLen == 0 {
		return gstTypes.ErrTxTrackingDataNotFound
	}
	currentTxGasTracking := currentBlockGasTracking.TxTrackingInfos[txsLen-1]
	currentBlockGasTracking.TxTrackingInfos[txsLen-1].ContractTrackingInfos = append(currentTxGasTracking.ContractTrackingInfos, &gstTypes.ContractGasTracking{
		Address:             contractAddress,
		GasConsumed:         gasUsed,
		Operation:           operation,
		IsEligibleForReward: isEligibleForReward,
	})
	bz, err = k.appCodec.MarshalBinaryBare(&currentBlockGasTracking)
	if err != nil {
		return err
	}

	gstKvStore.Set([]byte(gstTypes.CurrentBlockTrackingKey), bz)
	return nil
}
