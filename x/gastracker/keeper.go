package gastracker

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
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
	wasmTypes.ContractGasProcessor

	TrackNewTx(ctx sdk.Context, fee []*sdk.DecCoin, gasLimit uint64) error
	GetCurrentBlockTracking(ctx sdk.Context) (gstTypes.BlockGasTracking, error)
	GetCurrentTxTracking(ctx sdk.Context) (gstTypes.TransactionTracking, error)
	TrackNewBlock(ctx sdk.Context) error

	GetContractMetadata(ctx sdk.Context, address sdk.AccAddress) (gstTypes.ContractInstanceMetadata, error)
	AddPendingChangeForContractMetadata(ctx sdk.Context, sender sdk.AccAddress, address sdk.AccAddress, newMetadata gstTypes.ContractInstanceMetadata) error
	CommitPendingContractMetadata(ctx sdk.Context) (int, error)

	CreateOrMergeLeftOverRewardEntry(ctx sdk.Context, rewardAddress sdk.AccAddress, contractRewards sdk.DecCoins, leftOverThreshold uint64) (sdk.Coins, error)
	GetLeftOverRewardEntry(ctx sdk.Context, rewardAddress sdk.AccAddress) (gstTypes.LeftOverRewardEntry, error)

	SetParams(ctx sdk.Context, params gstTypes.Params)
	IsGasTrackingEnabled(ctx sdk.Context) bool
	IsDappInflationRewardsEnabled(ctx sdk.Context) bool
	IsGasRebateEnabled(ctx sdk.Context) bool
	IsGasRebateToUserEnabled(ctx sdk.Context) bool
	IsContractPremiumEnabled(ctx sdk.Context) bool
}

type Keeper struct {
	key              sdk.StoreKey
	appCodec         codec.Codec
	paramSpace       gstTypes.Subspace
	contractInfoView ContractInfoView
	wasmGasRegister  wasmkeeper.GasRegister
}

func NewGasTrackingKeeper(
	key sdk.StoreKey,
	appCodec codec.Codec,
	paramSpace paramsTypes.Subspace,
	contractInfoView ContractInfoView,
	gasRegister wasmkeeper.GasRegister,
) *Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(gstTypes.ParamKeyTable())
	}
	return &Keeper{key: key, appCodec: appCodec, paramSpace: paramSpace, contractInfoView: contractInfoView, wasmGasRegister: gasRegister}
}

func (k *Keeper) IngestGasRecord(ctx sdk.Context, records []wasmTypes.ContractGasRecord) error {
	if !k.IsGasTrackingEnabled(ctx) {
		return nil
	}

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
			fallthrough
		case wasmTypes.ContractOperationIbcPacketAck:
			fallthrough
		case wasmTypes.ContractOperationIbcPacketReceive:
			fallthrough
		case wasmTypes.ContractOperationIbcChannelClose:
			fallthrough
		case wasmTypes.ContractOperationIbcChannelOpen:
			fallthrough
		case wasmTypes.ContractOperationIbcChannelConnect:
			operation = gstTypes.ContractOperation_CONTRACT_OPERATION_IBC
		default:
			operation = gstTypes.ContractOperation_CONTRACT_OPERATION_UNSPECIFIED
		}

		if err := k.TrackContractGasUsage(ctx, contractAddress, k.wasmGasRegister.FromWasmVMGas(record.GasConsumed), operation, !contractMetadata.GasRebateToUser); err != nil {
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

	// We are pre-fetching the configuration so that
	// gas usage is similar across all conditions.
	isGasUpdateEnabled := k.IsGasRebateEnabled(ctx)
	isContractPremiumEnabled := k.IsContractPremiumEnabled(ctx)

	if isGasUpdateEnabled && contractMetadata.GasRebateToUser {
		updatedGas = (updatedGas * 50) / 100
	}

	if isContractPremiumEnabled && contractMetadata.CollectPremium {
		updatedGas = updatedGas + (updatedGas*contractMetadata.PremiumPercentageCharged)/100
	}

	return updatedGas, nil
}

func (k *Keeper) GetCurrentTxTracking(ctx sdk.Context) (gstTypes.TransactionTracking, error) {
	var txTrackingInfo gstTypes.TransactionTracking

	gstKvStore := ctx.KVStore(k.key)
	bz := gstKvStore.Get([]byte(gstTypes.CurrentBlockTrackingKey))
	if bz == nil {
		return txTrackingInfo, gstTypes.ErrBlockTrackingDataNotFound
	}
	var currentBlockGasTracking gstTypes.BlockGasTracking
	err := k.appCodec.Unmarshal(bz, &currentBlockGasTracking)
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
		err := k.appCodec.Unmarshal(bz, &rewardEntry)
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

	bz, err := k.appCodec.Marshal(&rewardEntry)
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

	err := k.appCodec.Unmarshal(bz, &rewardEntry)
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

	err := k.appCodec.Unmarshal(bz, &contractInstanceMetadata)
	return contractInstanceMetadata, err
}

func (k *Keeper) GetPendingContractMetadataChange(ctx sdk.Context, address sdk.AccAddress) (gstTypes.ContractInstanceMetadata, error) {
	gstKvStore := ctx.KVStore(k.key)

	var contractInstanceMetadata gstTypes.ContractInstanceMetadata

	bz := gstKvStore.Get(gstTypes.GetPendingContractInstanceMetadataKey(address.String()))
	if bz == nil {
		return contractInstanceMetadata, gstTypes.ErrContractInstanceMetadataNotFound
	}

	err := k.appCodec.Unmarshal(bz, &contractInstanceMetadata)
	return contractInstanceMetadata, err
}

func (k *Keeper) AddPendingChangeForContractMetadata(ctx sdk.Context, sender sdk.AccAddress, address sdk.AccAddress, newMetadata gstTypes.ContractInstanceMetadata) error {
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

	bz, err := k.appCodec.Marshal(&newMetadata)
	if err != nil {
		return err
	}

	gstKvStore.Set(gstTypes.GetPendingContractInstanceMetadataKey(address.String()), bz)
	return nil
}

func (k *Keeper) CommitPendingContractMetadata(ctx sdk.Context) (int, error) {
	gstKvStore := ctx.KVStore(k.key)
	keysToBeDeleted := make([][]byte, 0)

	iterator := gstKvStore.Iterator(
		[]byte(gstTypes.PendingContractInstanceMetadataKeyPrefix),
		sdk.PrefixEndBytes([]byte(gstTypes.PendingContractInstanceMetadataKeyPrefix)),
	)

	defer func() {
		for _, key := range keysToBeDeleted {
			gstKvStore.Delete(key)
		}
		iterator.Close()
	}()
	for ; iterator.Valid(); iterator.Next() {
		contractAddress := gstTypes.SplitContractAddressFromPendingMetadataKey(iterator.Key())
		bz := iterator.Value()
		gstKvStore.Set(gstTypes.GetContractInstanceMetadataKey(contractAddress), bz)
		keysToBeDeleted = append(keysToBeDeleted, iterator.Key())
	}

	return len(keysToBeDeleted), nil
}

func (k *Keeper) TrackNewBlock(ctx sdk.Context) error {
	gstKvStore := ctx.KVStore(k.key)
	bz, err := k.appCodec.Marshal(&gstTypes.BlockGasTracking{})
	if err != nil {
		return err
	}
	gstKvStore.Set([]byte(gstTypes.CurrentBlockTrackingKey), bz)
	return nil
}

func (k *Keeper) GetCurrentBlockTracking(ctx sdk.Context) (gstTypes.BlockGasTracking, error) {
	gstKvStore := ctx.KVStore(k.key)

	var currentBlockTracking gstTypes.BlockGasTracking
	bz := gstKvStore.Get([]byte(gstTypes.CurrentBlockTrackingKey))
	if bz == nil {
		return currentBlockTracking, gstTypes.ErrBlockTrackingDataNotFound
	}
	err := k.appCodec.Unmarshal(bz, &currentBlockTracking)
	return currentBlockTracking, err
}

func (k *Keeper) TrackNewTx(ctx sdk.Context, fee []*sdk.DecCoin, gasLimit uint64) error {
	gstKvStore := ctx.KVStore(k.key)

	var currentTxGasTracking gstTypes.TransactionTracking
	currentTxGasTracking.MaxContractRewards = fee
	currentTxGasTracking.MaxGasAllowed = gasLimit
	bz, err := k.appCodec.Marshal(&currentTxGasTracking)
	if err != nil {
		return err
	}

	bz = gstKvStore.Get([]byte(gstTypes.CurrentBlockTrackingKey))
	if bz == nil {
		return gstTypes.ErrBlockTrackingDataNotFound
	}
	var currentBlockTracking gstTypes.BlockGasTracking
	err = k.appCodec.Unmarshal(bz, &currentBlockTracking)
	if err != nil {
		return err
	}
	currentBlockTracking.TxTrackingInfos = append(currentBlockTracking.TxTrackingInfos, &currentTxGasTracking)
	bz, err = k.appCodec.Marshal(&currentBlockTracking)
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
	err := k.appCodec.Unmarshal(bz, &currentBlockGasTracking)
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
	bz, err = k.appCodec.Marshal(&currentBlockGasTracking)
	if err != nil {
		return err
	}

	gstKvStore.Set([]byte(gstTypes.CurrentBlockTrackingKey), bz)
	return nil
}
