package keeper

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	key              sdk.StoreKey
	cdc              codec.Codec
	paramSpace       gastracker.Subspace
	contractInfoView ContractInfoView
	wasmGasRegister  wasmkeeper.GasRegister

	mintKeeper MintKeeper
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
		wasmGasRegister:  gasRegister,
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

func (k Keeper) TrackNewBlock(ctx sdk.Context) error {
	gstKvStore := ctx.KVStore(k.key)
	bz, err := k.cdc.Marshal(&gastracker.BlockGasTracking{})
	if err != nil {
		return err
	}
	gstKvStore.Set([]byte(gastracker.CurrentBlockTrackingKey), bz)
	return nil
}

func (k Keeper) GetCurrentBlockTracking(ctx sdk.Context) (gastracker.BlockGasTracking, error) {
	gstKvStore := ctx.KVStore(k.key)

	var currentBlockTracking gastracker.BlockGasTracking
	bz := gstKvStore.Get([]byte(gastracker.CurrentBlockTrackingKey))
	if bz == nil {
		return currentBlockTracking, gastracker.ErrBlockTrackingDataNotFound
	}
	err := k.cdc.Unmarshal(bz, &currentBlockTracking)
	return currentBlockTracking, err
}

func (k Keeper) TrackNewTx(ctx sdk.Context, fee []sdk.DecCoin, gasLimit uint64) error {
	gstKvStore := ctx.KVStore(k.key)

	var currentTxGasTracking gastracker.TransactionTracking
	currentTxGasTracking.MaxContractRewards = fee
	currentTxGasTracking.MaxGasAllowed = gasLimit

	bz := gstKvStore.Get([]byte(gastracker.CurrentBlockTrackingKey))
	if bz == nil {
		return gastracker.ErrBlockTrackingDataNotFound
	}
	var currentBlockTracking gastracker.BlockGasTracking
	err := k.cdc.Unmarshal(bz, &currentBlockTracking)
	if err != nil {
		return err
	}
	currentBlockTracking.TxTrackingInfos = append(currentBlockTracking.TxTrackingInfos, currentTxGasTracking)
	bz, err = k.cdc.Marshal(&currentBlockTracking)
	if err != nil {
		return err
	}
	gstKvStore.Set([]byte(gastracker.CurrentBlockTrackingKey), bz)
	return nil
}

func (k Keeper) TrackContractGasUsage(ctx sdk.Context, contractAddress sdk.AccAddress, originalGas wasmTypes.GasConsumptionInfo, operation gastracker.ContractOperation) error {

	gstKvStore := ctx.KVStore(k.key)
	bz := gstKvStore.Get([]byte(gastracker.CurrentBlockTrackingKey))
	if bz == nil {
		return gastracker.ErrBlockTrackingDataNotFound
	}
	var currentBlockGasTracking gastracker.BlockGasTracking
	err := k.cdc.Unmarshal(bz, &currentBlockGasTracking)
	if err != nil {
		return err
	}

	txsLen := len(currentBlockGasTracking.TxTrackingInfos)
	if txsLen == 0 {
		return gastracker.ErrTxTrackingDataNotFound
	}
	currentTxGasTracking := currentBlockGasTracking.TxTrackingInfos[txsLen-1]
	currentBlockGasTracking.TxTrackingInfos[txsLen-1].ContractTrackingInfos = append(currentTxGasTracking.ContractTrackingInfos, gastracker.ContractGasTracking{
		Address:        contractAddress.String(),
		OriginalVmGas:  originalGas.VMGas,
		OriginalSdkGas: originalGas.SDKGas,
		Operation:      operation,
	})
	bz, err = k.cdc.Marshal(&currentBlockGasTracking)
	if err != nil {
		return err
	}

	gstKvStore.Set([]byte(gastracker.CurrentBlockTrackingKey), bz)
	return nil
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
