package keeper

import (
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	gastracker "github.com/archway-network/archway/x/gastracker"
	wasmBindingTypes "github.com/archway-network/archway/x/gastracker/wasmbinding/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramsTypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type ContractInfoView interface {
	GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmTypes.ContractInfo
}

type Keeper struct {
	WasmGasRegister wasmkeeper.GasRegister // can safely be exported since it's readonly.

	key              sdk.StoreKey
	cdc              codec.Codec
	paramSpace       gastracker.Subspace
	contractInfoView ContractInfoView
}

func NewGasTrackingKeeper(
	key sdk.StoreKey,
	appCodec codec.Codec,
	paramSpace paramsTypes.Subspace,
	contractInfoView ContractInfoView,
	gasRegister wasmkeeper.GasRegister,
) Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(gastracker.ParamKeyTable())
	}
	return Keeper{
		key:              key,
		cdc:              appCodec,
		paramSpace:       paramSpace,
		contractInfoView: contractInfoView,
		WasmGasRegister:  gasRegister,
	}
}

func (k *Keeper) SetContractInfoView(viewer ContractInfoView) {
	k.contractInfoView = viewer
}

func (k Keeper) GetContractMetadata(ctx sdk.Context, address sdk.AccAddress) (gastracker.ContractInstanceMetadata, error) {
	gstKvStore := ctx.KVStore(k.key)

	var contractInstanceMetadata gastracker.ContractInstanceMetadata

	bz := gstKvStore.Get(gastracker.GetContractInstanceMetadataKey(address.String()))
	if bz == nil {
		return contractInstanceMetadata, gastracker.ErrContractInstanceMetadataNotFound
	}

	err := k.cdc.Unmarshal(bz, &contractInstanceMetadata)
	return contractInstanceMetadata, err
}

func (k Keeper) GetPendingContractMetadataChange(ctx sdk.Context, address sdk.AccAddress) (gastracker.ContractInstanceMetadata, error) {
	gstKvStore := ctx.KVStore(k.key)

	var contractInstanceMetadata gastracker.ContractInstanceMetadata

	bz := gstKvStore.Get(gastracker.GetPendingContractInstanceMetadataKey(address.String()))
	if bz == nil {
		return contractInstanceMetadata, gastracker.ErrContractInstanceMetadataNotFound
	}

	err := k.cdc.Unmarshal(bz, &contractInstanceMetadata)
	return contractInstanceMetadata, err
}

func (k Keeper) AddPendingChangeForContractMetadata(ctx sdk.Context, sender sdk.AccAddress, address sdk.AccAddress, newMetadata gastracker.ContractInstanceMetadata) error {
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

	bz, err := k.cdc.Marshal(&newMetadata)
	if err != nil {
		return err
	}

	gstKvStore.Set(gastracker.GetPendingContractInstanceMetadataKey(address.String()), bz)
	return nil
}

// AddPendingChangeForContractMetadataByContract is called by the contract via custom WASM binding to add a pending change for the contract's metadata.
func (k Keeper) AddPendingChangeForContractMetadataByContract(ctx sdk.Context, contractAddr sdk.AccAddress, req wasmBindingTypes.UpdateMetadataRequest) error {
	// Input checks
	metadata, err := k.GetContractMetadata(ctx, contractAddr)
	if err != nil {
		// ErrContractInstanceMetadataNotFound is OK here, since we can't update a non-existing metadata
		return err
	}

	if metadata.DeveloperAddress != contractAddr.String() {
		return gastracker.ErrNoPermissionToSetMetadata
	}

	// Update
	if newAddr, isSet := req.GetDeveloperAddress(); isSet {
		metadata.DeveloperAddress = newAddr
	}

	if newAddr, isSet := req.GetRewardAddress(); isSet {
		metadata.RewardAddress = newAddr
	}

	// Set
	metadataBz, err := k.cdc.Marshal(&metadata)
	if err != nil {
		return sdkErrors.Wrapf(gastracker.ErrInternal, "metadata marshal: %v", err)
	}

	store := ctx.KVStore(k.key)
	metadataKey := gastracker.GetPendingContractInstanceMetadataKey(contractAddr.String())
	store.Set(metadataKey, metadataBz)

	return nil
}

func (k Keeper) CommitPendingContractMetadata(ctx sdk.Context) (int, error) {
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

func (k Keeper) TrackNewBlock(ctx sdk.Context) {
	// reset tx identifier
	k.ResetTxIdentifier(ctx)
	// delete tx tracking information
	store := prefix.NewStore(ctx.KVStore(k.key), append(gastracker.PrefixGasTrackingTxTracking, sdk.Uint64ToBigEndian(uint64(ctx.BlockHeight()))...))
	iter := store.Iterator(nil, nil)
	var keys [][]byte
	for ; iter.Valid(); iter.Next() {
		keys = append(keys, iter.Key())
	}

	for _, key := range keys {
		store.Delete(key)
	}
}

func (k Keeper) GetCurrentBlockTracking(ctx sdk.Context) gastracker.BlockGasTracking {
	store := prefix.NewStore(ctx.KVStore(k.key), gastracker.PrefixGasTrackingTxTracking)
	// we prefix over current block height
	iter := prefix.NewStore(store, sdk.Uint64ToBigEndian(uint64(ctx.BlockHeight()))).Iterator(nil, nil)
	defer iter.Close()
	var currentBlockTracking gastracker.BlockGasTracking

	for ; iter.Valid(); iter.Next() {
		v := gastracker.TransactionTracking{}
		k.cdc.MustUnmarshal(iter.Value(), &v)
		currentBlockTracking.TxTrackingInfos = append(currentBlockTracking.TxTrackingInfos, v)
	}

	return currentBlockTracking
}

func (k Keeper) TrackNewTx(ctx sdk.Context, fee []sdk.DecCoin, gasLimit uint64) {
	txIdentifier := k.GetAndIncreaseTxIdentifier(ctx)
	store := prefix.NewStore(ctx.KVStore(k.key), gastracker.PrefixGasTrackingTxTracking)

	var currentTxGasTracking gastracker.TransactionTracking
	currentTxGasTracking.MaxContractRewards = fee
	currentTxGasTracking.MaxGasAllowed = gasLimit

	store.Set(
		append(sdk.Uint64ToBigEndian(uint64(ctx.BlockHeight())), sdk.Uint64ToBigEndian(txIdentifier)...),
		k.cdc.MustMarshal(&currentTxGasTracking),
	)
}

func (k Keeper) TrackContractGasUsage(ctx sdk.Context, contractAddress sdk.AccAddress, originalGas wasmTypes.GasConsumptionInfo, operation gastracker.ContractOperation) {
	store := prefix.NewStore(ctx.KVStore(k.key), gastracker.PrefixGasTrackingTxTracking)
	key := append(sdk.Uint64ToBigEndian(uint64(ctx.BlockHeight())), sdk.Uint64ToBigEndian(k.GetCurrentTxIdentifier(ctx))...)

	bytes := store.Get(key)
	if bytes == nil {
		// panicking here since TrackNewTx must always be called before
		// we start appending operations to contracts.
		panic(fmt.Errorf("no gas tracking found for tx"))
	}
	var transactionTracking gastracker.TransactionTracking
	k.cdc.MustUnmarshal(bytes, &transactionTracking)

	transactionTracking.ContractTrackingInfos = append(transactionTracking.ContractTrackingInfos, gastracker.ContractGasTracking{
		Address:        contractAddress.String(),
		OriginalVmGas:  originalGas.VMGas,
		OriginalSdkGas: originalGas.SDKGas,
		Operation:      operation,
	})
	store.Set(key, k.cdc.MustMarshal(&transactionTracking))
}

// GetAndIncreaseTxIdentifier gets the current Tx identifier.
// Then increases the current tx identifier by 1.
// Assumes there is already a valid value saved in the store.
func (k Keeper) GetAndIncreaseTxIdentifier(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.key), gastracker.PrefixGasTrackingTxIdentifier)
	value := sdk.BigEndianToUint64(store.Get(gastracker.KeyTxIdentifier))
	store.Set(gastracker.KeyTxIdentifier, sdk.Uint64ToBigEndian(value+1))
	return value
}

// GetCurrentTxIdentifier gets the current Tx identifier.
// Contract: assumes GetAndIncreaseTxIdentifier was called
// at least once in this block.
func (k Keeper) GetCurrentTxIdentifier(ctx sdk.Context) uint64 {
	return sdk.BigEndianToUint64(prefix.NewStore(ctx.KVStore(k.key), gastracker.PrefixGasTrackingTxIdentifier).Get(gastracker.KeyTxIdentifier)) - 1
}

// ResetTxIdentifier resets the current Tx identifier to 0.
func (k Keeper) ResetTxIdentifier(ctx sdk.Context) {
	prefix.NewStore(ctx.KVStore(k.key), gastracker.PrefixGasTrackingTxIdentifier).Set(gastracker.KeyTxIdentifier, sdk.Uint64ToBigEndian(0))
}

// UpdateDappInflationaryRewards sets the current block inflationary rewards.
// Returns the inflationary rewards to be distributed.
func (k Keeper) UpdateDappInflationaryRewards(ctx sdk.Context, rewards sdk.Coin) {
	store := prefix.NewStore(ctx.KVStore(k.key), gastracker.PrefixDappBlockInflationaryRewards)

	store.Set(
		sdk.Uint64ToBigEndian(uint64(ctx.BlockHeight())),
		k.cdc.MustMarshal(&rewards),
	)
}

// GetDappInflationaryRewards returns the dApp inflationary rewards at the given height.
func (k Keeper) GetDappInflationaryRewards(ctx sdk.Context, height int64) (rewards sdk.DecCoin, err error) {
	store := prefix.NewStore(ctx.KVStore(k.key), gastracker.PrefixDappBlockInflationaryRewards)
	bytes := store.Get(sdk.Uint64ToBigEndian(uint64(height)))
	if bytes == nil {
		return rewards, gastracker.ErrDappInflationaryRewardRecordNotFound.Wrapf("height %d", height)
	}

	coin := sdk.Coin{}
	k.cdc.MustUnmarshal(bytes, &coin)
	return sdk.NewDecCoinFromCoin(coin), nil
}

// GetCurrentBlockDappInflationaryRewards returns the current block inflationary rewards.
func (k Keeper) GetCurrentBlockDappInflationaryRewards(ctx sdk.Context) (sdk.DecCoin, error) {
	return k.GetDappInflationaryRewards(ctx, ctx.BlockHeight())
}
