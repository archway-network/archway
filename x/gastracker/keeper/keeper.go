package keeper

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsTypes "github.com/cosmos/cosmos-sdk/x/params/types"

	gastracker "github.com/archway-network/archway/x/gastracker"
)

var _ GasTrackingKeeper = &Keeper{}
var _ wasmTypes.ContractGasProcessor = &Keeper{}

type ContractInfoView interface {
	GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmTypes.ContractInfo
}

type GasTrackingKeeper interface {
	wasmTypes.ContractGasProcessor

	TrackNewTx(ctx sdk.Context, fee []*sdk.DecCoin, remainingFeeCoins []*sdk.DecCoin, gasLimit uint64) error
	GetCurrentBlockTracking(ctx sdk.Context) (gastracker.BlockGasTracking, error)
	GetCurrentTxTracking(ctx sdk.Context) (gastracker.TransactionTracking, error)
	TrackNewBlock(ctx sdk.Context) error

	GetContractMetadata(ctx sdk.Context, address sdk.AccAddress) (gastracker.ContractInstanceMetadata, error)
	AddPendingChangeForContractMetadata(ctx sdk.Context, sender sdk.AccAddress, address sdk.AccAddress, newMetadata gastracker.ContractInstanceMetadata) error
	CommitPendingContractMetadata(ctx sdk.Context) (int, error)

	CreateOrMergeLeftOverRewardEntry(ctx sdk.Context, rewardAddress sdk.AccAddress, contractRewards sdk.DecCoins, leftOverThreshold uint64) (sdk.Coins, error)
	GetLeftOverRewardEntry(ctx sdk.Context, rewardAddress sdk.AccAddress) (gastracker.LeftOverRewardEntry, error)

	SetParams(ctx sdk.Context, params gastracker.Params)

	// IsGasTrackingEnabled gives a flag which describes whether gas tracking functionality is enabled or not
	IsGasTrackingEnabled(ctx sdk.Context) bool

	// IsDappInflationRewardsEnabled gives a flag which describes whether inflation reward is enabled or not
	IsDappInflationRewardsEnabled(ctx sdk.Context) bool

	// IsGasRebateToContractEnabled gives a flag which describes whether gas reward to contract is enabled or not
	IsGasRebateToContractEnabled(ctx sdk.Context) bool

	// IsGasRebateToUserEnabled gives a flag which describes whether gas reward to user is enabled or not
	IsGasRebateToUserEnabled(ctx sdk.Context) bool

	// IsContractPremiumEnabled gives a flag which describes whether contract premium is enabled or not
	IsContractPremiumEnabled(ctx sdk.Context) bool

	IsInflationRewardCapped(ctx sdk.Context) bool

	InflationRewardQuotaPercentage(ctx sdk.Context) uint64

	InflationRewardCapPercentage(ctx sdk.Context) uint64
}

type Keeper struct {
	key              sdk.StoreKey
	appCodec         codec.Codec
	paramSpace       gastracker.Subspace
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
		paramSpace = paramSpace.WithKeyTable(gastracker.ParamKeyTable())
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
		_, err = k.GetContractMetadata(ctx, contractAddress)
		switch err {
		case gastracker.ErrContractInstanceMetadataNotFound:
			contractMetadataExists = false
		case nil:
			contractMetadataExists = true
		default:
			return err
		}

		if !contractMetadataExists {
			continue
		}

		var operation gastracker.ContractOperation
		switch record.OperationId {
		case wasmTypes.ContractOperationQuery:
			operation = gastracker.ContractOperation_CONTRACT_OPERATION_QUERY
		case wasmTypes.ContractOperationInstantiate:
			operation = gastracker.ContractOperation_CONTRACT_OPERATION_INSTANTIATION
		case wasmTypes.ContractOperationExecute:
			operation = gastracker.ContractOperation_CONTRACT_OPERATION_EXECUTION
		case wasmTypes.ContractOperationMigrate:
			operation = gastracker.ContractOperation_CONTRACT_OPERATION_MIGRATE
		case wasmTypes.ContractOperationSudo:
			operation = gastracker.ContractOperation_CONTRACT_OPERATION_SUDO
		case wasmTypes.ContractOperationReply:
			operation = gastracker.ContractOperation_CONTRACT_OPERATION_REPLY
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
			operation = gastracker.ContractOperation_CONTRACT_OPERATION_IBC
		default:
			operation = gastracker.ContractOperation_CONTRACT_OPERATION_UNSPECIFIED
		}

		if err := k.TrackContractGasUsage(ctx, contractAddress, wasmTypes.GasConsumptionInfo{
			SDKGas: record.OriginalGas.SDKGas,
			VMGas:  k.wasmGasRegister.FromWasmVMGas(record.OriginalGas.VMGas),
		}, operation); err != nil {
			return err
		}
	}

	return nil
}

func (k *Keeper) GetGasCalculationFn(ctx sdk.Context, contractAddress string) (func(operationId uint64, gasInfo wasmTypes.GasConsumptionInfo) wasmTypes.GasConsumptionInfo, error) {
	var contractMetadataExists bool

	passthroughFn := func(operationId uint64, gasConsumptionInfo wasmTypes.GasConsumptionInfo) wasmTypes.GasConsumptionInfo {
		return gasConsumptionInfo
	}

	doNotUse := func(operationId uint64, _ wasmTypes.GasConsumptionInfo) wasmTypes.GasConsumptionInfo {
		panic("do not use this function")
	}

	contractAddr, err := sdk.AccAddressFromBech32(contractAddress)
	if err != nil {
		return doNotUse, err
	}

	contractMetadata, err := k.GetContractMetadata(ctx, contractAddr)
	switch err {
	case gastracker.ErrContractInstanceMetadataNotFound:
		contractMetadataExists = false
	case nil:
		contractMetadataExists = true
	default:
		return doNotUse, err
	}

	if !contractMetadataExists {
		return passthroughFn, nil
	}

	// We are pre-fetching the configuration so that
	// gas usage is similar across all conditions.
	isGasRebateToUserEnabled := k.IsGasRebateToUserEnabled(ctx)
	isContractPremiumEnabled := k.IsContractPremiumEnabled(ctx)
	isGasTrackingEnabled := k.IsGasTrackingEnabled(ctx)

	return func(operationId uint64, gasConsumptionInfo wasmTypes.GasConsumptionInfo) wasmTypes.GasConsumptionInfo {
		if !isGasTrackingEnabled {
			return gasConsumptionInfo
		}

		if isGasRebateToUserEnabled && contractMetadata.GasRebateToUser {
			updatedGas := wasmTypes.GasConsumptionInfo{
				SDKGas: (gasConsumptionInfo.SDKGas * 50) / 100,
				VMGas:  (gasConsumptionInfo.VMGas * 50) / 100,
			}
			return updatedGas
		} else if isContractPremiumEnabled && contractMetadata.CollectPremium {
			updatedGas := wasmTypes.GasConsumptionInfo{
				SDKGas: gasConsumptionInfo.SDKGas + (gasConsumptionInfo.SDKGas*contractMetadata.PremiumPercentageCharged)/100,
				VMGas:  gasConsumptionInfo.VMGas + (gasConsumptionInfo.VMGas*contractMetadata.PremiumPercentageCharged)/100,
			}
			return updatedGas
		} else {
			return gasConsumptionInfo
		}
	}, nil
}

func (k *Keeper) CalculateUpdatedGas(ctx sdk.Context, record wasmTypes.ContractGasRecord) (wasmTypes.GasConsumptionInfo, error) {
	gasCalcFn, err := k.GetGasCalculationFn(ctx, record.ContractAddress)
	if err != nil {
		return wasmTypes.GasConsumptionInfo{}, nil
	}

	return gasCalcFn(record.OperationId, record.OriginalGas), nil
}

func (k *Keeper) GetCurrentTxTracking(ctx sdk.Context) (gastracker.TransactionTracking, error) {
	var txTrackingInfo gastracker.TransactionTracking

	gstKvStore := ctx.KVStore(k.key)
	bz := gstKvStore.Get([]byte(gastracker.CurrentBlockTrackingKey))
	if bz == nil {
		return txTrackingInfo, gastracker.ErrBlockTrackingDataNotFound
	}
	var currentBlockGasTracking gastracker.BlockGasTracking
	err := k.appCodec.Unmarshal(bz, &currentBlockGasTracking)
	if err != nil {
		return txTrackingInfo, err
	}

	txsLen := len(currentBlockGasTracking.TxTrackingInfos)
	if txsLen == 0 {
		return txTrackingInfo, gastracker.ErrTxTrackingDataNotFound
	}

	txTrackingInfo = *currentBlockGasTracking.TxTrackingInfos[len(currentBlockGasTracking.TxTrackingInfos)-1]
	return txTrackingInfo, nil
}

func (k *Keeper) CreateOrMergeLeftOverRewardEntry(ctx sdk.Context, rewardAddress sdk.AccAddress, contractRewards sdk.DecCoins, leftOverThreshold uint64) (sdk.Coins, error) {
	contractRewards = contractRewards.Sort()

	gstKvStore := ctx.KVStore(k.key)

	var rewardEntry gastracker.LeftOverRewardEntry
	var updatedRewards sdk.DecCoins
	var rewardsToBeDistributed sdk.Coins

	bz := gstKvStore.Get(gastracker.GetRewardEntryKey(rewardAddress.String()))
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

	leftOverDec := sdk.NewDecFromBigInt(gastracker.ConvertUint64ToBigInt(leftOverThreshold))

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

	gstKvStore.Set(gastracker.GetRewardEntryKey(rewardAddress.String()), bz)
	return rewardsToBeDistributed, nil
}

// Since we can only transfer integer numbers
// and rewards can be floating point numbers,
// we accumulate all the rewards and once it reaches to
// an integer number, we pay the integer part and
// keep the 0.x amount as left over to be paid later
func (k *Keeper) GetLeftOverRewardEntry(ctx sdk.Context, rewardAddress sdk.AccAddress) (gastracker.LeftOverRewardEntry, error) {
	gstKvStore := ctx.KVStore(k.key)

	var rewardEntry gastracker.LeftOverRewardEntry

	bz := gstKvStore.Get(gastracker.GetRewardEntryKey(rewardAddress.String()))
	if bz == nil {
		return rewardEntry, gastracker.ErrRewardEntryNotFound
	}

	err := k.appCodec.Unmarshal(bz, &rewardEntry)
	if err != nil {
		return rewardEntry, err
	}

	return rewardEntry, nil
}

func (k *Keeper) GetContractMetadata(ctx sdk.Context, address sdk.AccAddress) (gastracker.ContractInstanceMetadata, error) {
	gstKvStore := ctx.KVStore(k.key)

	var contractInstanceMetadata gastracker.ContractInstanceMetadata

	bz := gstKvStore.Get(gastracker.GetContractInstanceMetadataKey(address.String()))
	if bz == nil {
		return contractInstanceMetadata, gastracker.ErrContractInstanceMetadataNotFound
	}

	err := k.appCodec.Unmarshal(bz, &contractInstanceMetadata)
	return contractInstanceMetadata, err
}

func (k *Keeper) GetPendingContractMetadataChange(ctx sdk.Context, address sdk.AccAddress) (gastracker.ContractInstanceMetadata, error) {
	gstKvStore := ctx.KVStore(k.key)

	var contractInstanceMetadata gastracker.ContractInstanceMetadata

	bz := gstKvStore.Get(gastracker.GetPendingContractInstanceMetadataKey(address.String()))
	if bz == nil {
		return contractInstanceMetadata, gastracker.ErrContractInstanceMetadataNotFound
	}

	err := k.appCodec.Unmarshal(bz, &contractInstanceMetadata)
	return contractInstanceMetadata, err
}

func (k *Keeper) AddPendingChangeForContractMetadata(ctx sdk.Context, sender sdk.AccAddress, address sdk.AccAddress, newMetadata gastracker.ContractInstanceMetadata) error {
	gstKvStore := ctx.KVStore(k.key)

	contractInfo := k.contractInfoView.GetContractInfo(ctx, address)
	if contractInfo == nil {
		return gastracker.ErrContractInfoNotFound
	}

	contractMetadataExists := true
	instanceMetadata, err := k.GetContractMetadata(ctx, address)
	if err != nil {
		if err == gastracker.ErrContractInstanceMetadataNotFound {
			contractMetadataExists = false
		} else {
			return err
		}
	}

	if len(newMetadata.DeveloperAddress) == 0 {
		if contractMetadataExists {
			newMetadata.DeveloperAddress = instanceMetadata.DeveloperAddress
		} else {
			return gastracker.ErrInvalidSetContractMetadataRequest
		}
	}

	if len(newMetadata.RewardAddress) == 0 {
		if contractMetadataExists {
			newMetadata.RewardAddress = instanceMetadata.RewardAddress
		} else {
			return gastracker.ErrInvalidSetContractMetadataRequest
		}
	}

	if contractMetadataExists {
		if sender.String() != instanceMetadata.DeveloperAddress {
			return gastracker.ErrNoPermissionToSetMetadata
		}
	} else {
		if sender.String() != contractInfo.Admin {
			return gastracker.ErrNoPermissionToSetMetadata
		}
	}

	bz, err := k.appCodec.Marshal(&newMetadata)
	if err != nil {
		return err
	}

	gstKvStore.Set(gastracker.GetPendingContractInstanceMetadataKey(address.String()), bz)
	return nil
}

func (k *Keeper) CommitPendingContractMetadata(ctx sdk.Context) (int, error) {
	gstKvStore := ctx.KVStore(k.key)
	keysToBeDeleted := make([][]byte, 0)

	iterator := gstKvStore.Iterator(
		[]byte(gastracker.PendingContractInstanceMetadataKeyPrefix),
		sdk.PrefixEndBytes([]byte(gastracker.PendingContractInstanceMetadataKeyPrefix)),
	)

	defer func() {
		for _, key := range keysToBeDeleted {
			gstKvStore.Delete(key)
		}
		iterator.Close()
	}()
	for ; iterator.Valid(); iterator.Next() {
		contractAddress := gastracker.SplitContractAddressFromPendingMetadataKey(iterator.Key())
		bz := iterator.Value()
		gstKvStore.Set(gastracker.GetContractInstanceMetadataKey(contractAddress), bz)
		keysToBeDeleted = append(keysToBeDeleted, iterator.Key())
	}

	return len(keysToBeDeleted), nil
}

func (k *Keeper) TrackNewBlock(ctx sdk.Context) error {
	gstKvStore := ctx.KVStore(k.key)
	bz, err := k.appCodec.Marshal(&gastracker.BlockGasTracking{})
	if err != nil {
		return err
	}
	gstKvStore.Set([]byte(gastracker.CurrentBlockTrackingKey), bz)
	return nil
}

func (k *Keeper) GetCurrentBlockTracking(ctx sdk.Context) (gastracker.BlockGasTracking, error) {
	gstKvStore := ctx.KVStore(k.key)

	var currentBlockTracking gastracker.BlockGasTracking
	bz := gstKvStore.Get([]byte(gastracker.CurrentBlockTrackingKey))
	if bz == nil {
		return currentBlockTracking, gastracker.ErrBlockTrackingDataNotFound
	}
	err := k.appCodec.Unmarshal(bz, &currentBlockTracking)
	return currentBlockTracking, err
}

func (k *Keeper) TrackNewTx(ctx sdk.Context, fee []*sdk.DecCoin, remainingFeeCoins []*sdk.DecCoin, gasLimit uint64) error {
	gstKvStore := ctx.KVStore(k.key)

	var currentTxGasTracking gastracker.TransactionTracking
	currentTxGasTracking.MaxContractRewards = fee
	currentTxGasTracking.MaxGasAllowed = gasLimit
	currentTxGasTracking.RemainingFee = remainingFeeCoins
	bz, err := k.appCodec.Marshal(&currentTxGasTracking)
	if err != nil {
		return err
	}

	bz = gstKvStore.Get([]byte(gastracker.CurrentBlockTrackingKey))
	if bz == nil {
		return gastracker.ErrBlockTrackingDataNotFound
	}
	var currentBlockTracking gastracker.BlockGasTracking
	err = k.appCodec.Unmarshal(bz, &currentBlockTracking)
	if err != nil {
		return err
	}
	currentBlockTracking.TxTrackingInfos = append(currentBlockTracking.TxTrackingInfos, &currentTxGasTracking)
	bz, err = k.appCodec.Marshal(&currentBlockTracking)
	if err != nil {
		return err
	}
	gstKvStore.Set([]byte(gastracker.CurrentBlockTrackingKey), bz)
	return nil
}

func (k *Keeper) TrackContractGasUsage(ctx sdk.Context, contractAddress sdk.AccAddress, originalGas wasmTypes.GasConsumptionInfo, operation gastracker.ContractOperation) error {
	gstKvStore := ctx.KVStore(k.key)
	bz := gstKvStore.Get([]byte(gastracker.CurrentBlockTrackingKey))
	if bz == nil {
		return gastracker.ErrBlockTrackingDataNotFound
	}
	var currentBlockGasTracking gastracker.BlockGasTracking
	err := k.appCodec.Unmarshal(bz, &currentBlockGasTracking)
	if err != nil {
		return err
	}

	txsLen := len(currentBlockGasTracking.TxTrackingInfos)
	if txsLen == 0 {
		return gastracker.ErrTxTrackingDataNotFound
	}
	currentTxGasTracking := currentBlockGasTracking.TxTrackingInfos[txsLen-1]
	currentBlockGasTracking.TxTrackingInfos[txsLen-1].ContractTrackingInfos = append(currentTxGasTracking.ContractTrackingInfos, &gastracker.ContractGasTracking{
		Address:        contractAddress.String(),
		OriginalVmGas:  originalGas.VMGas,
		OriginalSdkGas: originalGas.SDKGas,
		Operation:      operation,
	})
	bz, err = k.appCodec.Marshal(&currentBlockGasTracking)
	if err != nil {
		return err
	}

	gstKvStore.Set([]byte(gastracker.CurrentBlockTrackingKey), bz)
	return nil
}
