package gastracker

import (
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsTypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ GasTrackingKeeper = &Keeper{}
var _ wasmTypes.ContractGasProcessor = &Keeper{}

type ContractInfoView interface {
	GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmTypes.ContractInfo
}

type GasTrackingKeeper interface {
	TrackNewTx(ctx sdk.Context, fee []*sdk.DecCoin, gasLimit uint64) error
	TrackContractGasUsage(ctx sdk.Context, contractAddress sdk.AccAddress, gasUsed uint64, operation gstTypes.ContractOperation, isEligibleForReward bool) error
	GetCurrentBlockTrackingInfo(ctx sdk.Context) (gstTypes.BlockGasTracking, error)
	GetCurrentTxTrackingInfo(ctx sdk.Context) (gstTypes.TransactionTracking, error)
	TrackNewBlock(ctx sdk.Context) error
	SetContractMetadata(ctx sdk.Context, admin sdk.AccAddress, address sdk.AccAddress, metadata gstTypes.ContractInstanceMetadata) error
	GetContractMetadata(ctx sdk.Context, address sdk.AccAddress) (gstTypes.ContractInstanceMetadata, error)

	CreateOrMergeLeftOverRewardEntry(ctx sdk.Context, rewardAddress sdk.AccAddress, contractRewards sdk.DecCoins, leftOverThreshold uint64) (sdk.Coins, error)
	GetLeftOverRewardEntry(ctx sdk.Context, rewardAddress sdk.AccAddress) (gstTypes.LeftOverRewardEntry, error)

	SetParams(ctx sdk.Context, params gstTypes.Params)
	IsGasTrackingEnabled(ctx sdk.Context) bool
	IsDappInflationRewardsEnabled(ctx sdk.Context) bool
	IsGasRebateEnabled(ctx sdk.Context) bool
	IsGasRebateToUserEnabled(ctx sdk.Context) bool
	IsContractPremiumEnabled(ctx sdk.Context) bool

	IngestGasRecord(ctx sdk.Context, records []wasmTypes.ContractGasRecord) error
	CalculateUpdatedGas(ctx sdk.Context, record wasmTypes.ContractGasRecord) (uint64, error)
}

type Keeper struct {
	key              sdk.StoreKey
	appCodec         codec.Marshaler
	paramSpace       gstTypes.Subspace
	contractInfoView ContractInfoView
}

func (k *Keeper) IngestGasRecord(ctx sdk.Context, records []wasmTypes.ContractGasRecord) error {
	for _, record := range records {
		contractAddress, err := sdk.AccAddressFromBech32(record.ContractAddress)
		if err != nil {
			return err
		}

		var contractMetadataExists bool
		contractMetadata, err := k.GetContractMetadata(ctx, contractAddress)
		switch err {
		case gstTypes.ErrContractInstanceMetadataNotFound:
			contractMetadataExists = false
		case nil:
			contractMetadataExists = true
		default:
			return err
		}

		if !contractMetadataExists {
			continue
		}

		var operation gstTypes.ContractOperation
		switch record.OperationId {
		case wasmTypes.ContractOperationQuery:
			operation = gstTypes.ContractOperation_CONTRACT_OPERATION_QUERY
		case wasmTypes.ContractOperationInstantiate:
			operation = gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION
		case wasmTypes.ContractOperationExecute:
			operation = gstTypes.ContractOperation_CONTRACT_OPERATION_EXECUTION
		case wasmTypes.ContractOperationMigrate:
			operation = gstTypes.ContractOperation_CONTRACT_OPERATION_MIGRATE
		case wasmTypes.ContractOperationSudo:
			operation = gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO
		case wasmTypes.ContractOperationReply:
			operation = gstTypes.ContractOperation_CONTRACT_OPERATION_REPLY
		case wasmTypes.ContractOperationIbcPacketTimeout:
		case wasmTypes.ContractOperationIbcPacketAck:
		case wasmTypes.ContractOperationIbcPacketReceive:
		case wasmTypes.ContractOperationIbcChannelClose:
		case wasmTypes.ContractOperationIbcChannelOpen:
		case wasmTypes.ContractOperationIbcChannelConnect:
			operation = gstTypes.ContractOperation_CONTRACT_OPERATION_IBC
		default:
			operation = gstTypes.ContractOperation_CONTRACT_OPERATION_UNSPECIFIED
		}

		if err := k.TrackContractGasUsage(ctx, contractAddress, record.GasConsumed, operation, !contractMetadata.GasRebateToUser); err != nil {
			return err
		}
	}

	return nil
}

func (k *Keeper) CalculateUpdatedGas(ctx sdk.Context, record wasmTypes.ContractGasRecord) (uint64, error) {
	var contractMetadataExists bool

	contractAddress, err := sdk.AccAddressFromBech32(record.ContractAddress)
	if err != nil {
		return 0, err
	}

	contractMetadata, err := k.GetContractMetadata(ctx, contractAddress)
	switch err {
	case gstTypes.ErrContractInstanceMetadataNotFound:
		contractMetadataExists = false
	case nil:
		contractMetadataExists = true
	default:
		return 0, err
	}

	updatedGas := record.GasConsumed

	if !contractMetadataExists {
		return updatedGas, nil
	}

	if contractMetadata.GasRebateToUser {
		updatedGas = (updatedGas * 50) / 100
	}

	if contractMetadata.CollectPremium {
		updatedGas = updatedGas + (updatedGas*contractMetadata.PremiumPercentageCharged)/100
	}

	return updatedGas, nil
}

func (k *Keeper) GetCurrentTxTrackingInfo(ctx sdk.Context) (gstTypes.TransactionTracking, error) {
	var txTrackingInfo gstTypes.TransactionTracking

	gstKvStore := ctx.KVStore(k.key)
	bz := gstKvStore.Get([]byte(gstTypes.CurrentBlockTrackingKey))
	if bz == nil {
		return txTrackingInfo, gstTypes.ErrBlockTrackingDataNotFound
	}
	var currentBlockGasTracking gstTypes.BlockGasTracking
	err := k.appCodec.UnmarshalBinaryBare(bz, &currentBlockGasTracking)
	if err != nil {
		return txTrackingInfo, err
	}

	txsLen := len(currentBlockGasTracking.TxTrackingInfos)
	if txsLen == 0 {
		return txTrackingInfo, gstTypes.ErrTxTrackingDataNotFound
	}

	txTrackingInfo = *currentBlockGasTracking.TxTrackingInfos[len(currentBlockGasTracking.TxTrackingInfos)-1]
	return txTrackingInfo, nil
}

func (k *Keeper) CreateOrMergeLeftOverRewardEntry(ctx sdk.Context, rewardAddress sdk.AccAddress, contractRewards sdk.DecCoins, leftOverThreshold uint64) (sdk.Coins, error) {
	contractRewards = contractRewards.Sort()

	gstKvStore := ctx.KVStore(k.key)

	var rewardEntry gstTypes.LeftOverRewardEntry
	var updatedRewards sdk.DecCoins
	var rewardsToBeDistributed sdk.Coins

	bz := gstKvStore.Get(gstTypes.GetRewardEntryKey(rewardAddress.String()))
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

	gstKvStore.Set(gstTypes.GetRewardEntryKey(rewardAddress.String()), bz)
	return rewardsToBeDistributed, nil
}

// Since we can only transfer integer numbers
// and rewards can be floating point numbers,
// we accumulate all the rewards and once it reaches to
// an integer number, we pay the integer part and
// keep the 0.x amount as left over to be paid later
func (k *Keeper) GetLeftOverRewardEntry(ctx sdk.Context, rewardAddress sdk.AccAddress) (gstTypes.LeftOverRewardEntry, error) {
	gstKvStore := ctx.KVStore(k.key)

	var rewardEntry gstTypes.LeftOverRewardEntry

	bz := gstKvStore.Get(gstTypes.GetRewardEntryKey(rewardAddress.String()))
	if bz == nil {
		return rewardEntry, gstTypes.ErrRewardEntryNotFound
	}

	err := k.appCodec.UnmarshalBinaryBare(bz, &rewardEntry)
	if err != nil {
		return rewardEntry, err
	}

	return rewardEntry, nil
}

func (k *Keeper) GetContractMetadata(ctx sdk.Context, address sdk.AccAddress) (gstTypes.ContractInstanceMetadata, error) {
	gstKvStore := ctx.KVStore(k.key)

	var contractInstanceMetadata gstTypes.ContractInstanceMetadata

	bz := gstKvStore.Get(gstTypes.GetContractInstanceMetadataKey(address.String()))
	if bz == nil {
		return contractInstanceMetadata, gstTypes.ErrContractInstanceMetadataNotFound
	}

	err := k.appCodec.UnmarshalBinaryBare(bz, &contractInstanceMetadata)
	return contractInstanceMetadata, err
}

// SetContractMetadata checks that the address that requested to add new metadata is admin of the contract or if admin is cleared the
// developer field of the contract metadata. \
// The developer field is set the first time when this contract metadata is set.
// Next time, only the admin (if not cleared) or the developer can change the metadata.
func (k *Keeper) SetContractMetadata(ctx sdk.Context, sender sdk.AccAddress, address sdk.AccAddress, newMetadata gstTypes.ContractInstanceMetadata) error {
	gstKvStore := ctx.KVStore(k.key)

	contractInfo := k.contractInfoView.GetContractInfo(ctx, address)
	if contractInfo == nil {
		return gstTypes.ErrContractInfoNotFound
	}

	contractMetadataExists := true
	instanceMetadata, err := k.GetContractMetadata(ctx, address)
	if err != nil {
		if err == gstTypes.ErrContractInstanceMetadataNotFound {
			contractMetadataExists = false
		} else {
			return err
		}
	}

	if len(newMetadata.DeveloperAddress) == 0 {
		if contractMetadataExists {
			newMetadata.DeveloperAddress = instanceMetadata.DeveloperAddress
		} else {
			return gstTypes.ErrInvalidSetContractMetadataRequest
		}
	}

	if len(newMetadata.RewardAddress) == 0 {
		if contractMetadataExists {
			newMetadata.RewardAddress = instanceMetadata.RewardAddress
		} else {
			return gstTypes.ErrInvalidSetContractMetadataRequest
		}
	}

	if contractMetadataExists {
		if sender.String() != instanceMetadata.DeveloperAddress {
			return gstTypes.ErrNoPermissionToSetMetadata
		}
	} else {
		if sender.String() != contractInfo.Admin {
			return gstTypes.ErrNoPermissionToSetMetadata
		}
	}

	bz, err := k.appCodec.MarshalBinaryBare(&newMetadata)
	if err != nil {
		return err
	}

	gstKvStore.Set(gstTypes.GetContractInstanceMetadataKey(address.String()), bz)
	return nil
}

func NewGasTrackingKeeper(key sdk.StoreKey, appCodec codec.Marshaler, paramSpace paramsTypes.Subspace, contractInfoView ContractInfoView) *Keeper {
	return &Keeper{key: key, appCodec: appCodec, paramSpace: paramSpace, contractInfoView: contractInfoView}
}

func (k *Keeper) TrackNewBlock(ctx sdk.Context) error {
	gstKvStore := ctx.KVStore(k.key)
	bz, err := k.appCodec.MarshalBinaryBare(&gstTypes.BlockGasTracking{})
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
	if bz == nil {
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

func (k *Keeper) TrackContractGasUsage(ctx sdk.Context, contractAddress sdk.AccAddress, gasUsed uint64, operation gstTypes.ContractOperation, isEligibleForReward bool) error {
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
		Address:             contractAddress.String(),
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
