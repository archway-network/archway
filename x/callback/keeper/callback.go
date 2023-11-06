package keeper

import (
	"github.com/archway-network/archway/x/callback/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetAllCallbacks lists all the pending callbacks
func (k Keeper) GetAllCallbacks(ctx sdk.Context) (callbacks []types.Callback) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.CallbackKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var c types.Callback
		k.cdc.MustUnmarshal(iterator.Value(), &c)
		callbacks = append(callbacks, c)
	}

	return callbacks
}

// GetCallbacksByHeight returns the callbacks registered for the given height
func (k Keeper) GetCallbacksByHeight(ctx sdk.Context, height int64) (callbacks []types.Callback) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetCallbacksByHeightKey(height))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var c types.Callback
		k.cdc.MustUnmarshal(iterator.Value(), &c)
		callbacks = append(callbacks, c)
	}

	return callbacks
}

// ExistsCallback returns true if the callback exists for height with same contract address and same job id
func (k Keeper) ExistsCallback(ctx sdk.Context, height int64, contractAddress sdk.AccAddress, jobID uint64) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetCallbackByFullyQualifiedKey(height, contractAddress, jobID)
	return store.Has(key)
}

// DeleteCallback deletes a callback given the height, contract address and job id
func (k Keeper) DeleteCallback(ctx sdk.Context, height int64, contractAddress sdk.AccAddress, jobID uint64, sender string) error {
	// If callback delete is requested by someone who is not authorized, return error
	if !isAuthorizedToModify(ctx, k, height, contractAddress, sender) {
		return types.ErrUnauthorized
	}
	// If a callback with same job id does not exist, return error
	if k.ExistsCallback(ctx, height, contractAddress, jobID) {
		return types.ErrCallbackJobIDDoesNotExists
	}
	store := ctx.KVStore(k.storeKey)
	key := types.GetCallbackByFullyQualifiedKey(height, contractAddress, jobID)
	store.Delete(key)
	return nil
}

// SaveCallback saves a callback given the height, contract address and job id and callback data
func (k Keeper) SaveCallback(ctx sdk.Context, height int64, contractAddress sdk.AccAddress, jobID uint64, callback types.Callback) error {
	// If contract with given address does not exist, return error
	if !k.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		return types.ErrContractNotFound
	}
	// If callback is requested by someone which is not authorized, return error
	if !isAuthorizedToModify(ctx, k, height, contractAddress, callback.ReservedBy) {
		return types.ErrUnauthorized
	}
	// If a callback with same job id exists at same height, return error
	if k.ExistsCallback(ctx, height, contractAddress, jobID) {
		return types.ErrCallbackJobIDExists
	}
	// If callback is requested for height in the past or present, return error
	if height <= ctx.BlockHeight() {
		return types.ErrCallbackHeightNotinFuture
	}
	store := ctx.KVStore(k.storeKey)
	key := types.GetCallbackByFullyQualifiedKey(height, contractAddress, jobID)
	bz, err := k.cdc.Marshal(&callback)
	if err != nil {
		return nil
	}
	store.Set(key, bz)
	return nil
}

func isAuthorizedToModify(ctx sdk.Context, k Keeper, height int64, contractAddress sdk.AccAddress, sender string) bool {
	if sender == contractAddress.String() { // A contract can modify its own callbacks
		return true
	}

	contractInfo := k.wasmKeeper.GetContractInfo(ctx, contractAddress)
	if sender == contractInfo.Admin { // Admin of the contract can modify its callbacks
		return true
	}

	contractMetadata := k.rewardsKeepers.GetContractMetadata(ctx, contractAddress)
	if sender == contractMetadata.OwnerAddress { // Owner of the contract can modify its callbacks
		return true
	}

	return false
}
