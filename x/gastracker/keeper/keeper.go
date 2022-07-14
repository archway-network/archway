package keeper

import (
	"fmt"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramsTypes "github.com/cosmos/cosmos-sdk/x/params/types"

	gastracker "github.com/archway-network/archway/x/gastracker"
)

type ContractInfoView interface {
	GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmTypes.ContractInfo
}

type MintKeeper interface {
	GetMinter(ctx sdk.Context) minttypes.Minter
	GetParams(ctx sdk.Context) minttypes.Params
}

type Keeper struct {
	WasmGasRegister wasmkeeper.GasRegister // can safely be exported since it's readonly.

	key              sdk.StoreKey
	cdc              codec.Codec
	paramSpace       gastracker.Subspace
	contractInfoView ContractInfoView
	mintKeeper       MintKeeper
}

func NewGasTrackingKeeper(
	key sdk.StoreKey,
	appCodec codec.Codec,
	paramSpace paramsTypes.Subspace,
	contractInfoView ContractInfoView,
	gasRegister wasmkeeper.GasRegister,
	mintKeeper MintKeeper,
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
		mintKeeper:       mintKeeper,
	}
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
func (k Keeper) UpdateDappInflationaryRewards(ctx sdk.Context, params gastracker.Params) (rewards sdk.DecCoin) {
	store := prefix.NewStore(ctx.KVStore(k.key), gastracker.PrefixDappBlockInflationaryRewards)

	// gets total inflationary rewards
	totalInflationaryRewards := k.mintKeeper.
		GetMinter(ctx).
		BlockProvision(k.mintKeeper.GetParams(ctx))

	dappInflationaryRewards := sdk.NewDecCoinFromDec(
		totalInflationaryRewards.Denom,
		totalInflationaryRewards.Amount.ToDec().Mul(params.DappInflationRewardsRatio),
	)

	store.Set(
		sdk.Uint64ToBigEndian(uint64(ctx.BlockHeight())),
		k.cdc.MustMarshal(&dappInflationaryRewards),
	)

	return dappInflationaryRewards
}

// GetDappInflationaryRewards returns the dApp inflationary rewards at the given height.
func (k Keeper) GetDappInflationaryRewards(ctx sdk.Context, height int64) (rewards sdk.DecCoin, err error) {
	store := prefix.NewStore(ctx.KVStore(k.key), gastracker.PrefixDappBlockInflationaryRewards)
	bytes := store.Get(sdk.Uint64ToBigEndian(uint64(height)))
	if bytes == nil {
		return rewards, gastracker.ErrDappInflationaryRewardRecordNotFound.Wrapf("height %d", height)
	}

	k.cdc.MustUnmarshal(bytes, &rewards)
	return rewards, nil
}
