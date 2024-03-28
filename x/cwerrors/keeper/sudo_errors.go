package keeper

import (
	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/cwerrors/types"
)

// SetError stores a sudo error and queues it for deletion after a certain block height
func (k Keeper) SetError(ctx sdk.Context, sudoErr types.SudoError) error {
	// Ensure error is valid
	if err := sudoErr.Validate(); err != nil {
		return err
	}
	contractAddr := sdk.MustAccAddressFromBech32(sudoErr.ContractAddress)
	// Ensure contract exists
	if !k.wasmKeeper.HasContractInfo(ctx, contractAddr) {
		return types.ErrContractNotFound
	}

	if k.HasSubscription(ctx, contractAddr) {
		// If contract has subscription, store the error in the transient store to be executed as error callback
		return k.storeErrorCallback(ctx, sudoErr)
	} else {
		// for contracts which dont have an error subscription, store the error in state to be deleted after a set height
		return k.StoreErrorInState(ctx, contractAddr, sudoErr)
	}
}

// StoreErrorInState stores the error in the state and queues it for deletion after a certain block height
func (k Keeper) StoreErrorInState(ctx sdk.Context, contractAddr sdk.AccAddress, sudoErr types.SudoError) error {
	// just a unique identifier for the error
	errorID, err := k.ErrorID.Next(ctx)
	if err != nil {
		return err
	}

	// Associate the error with the contract
	if err = k.ContractErrors.Set(ctx, collections.Join(contractAddr.Bytes(), errorID), errorID); err != nil {
		return err
	}

	// Store when the error should be deleted
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	deletionHeight := ctx.BlockHeight() + params.ErrorStoredTime
	if err = k.DeletionBlocks.Set(ctx, collections.Join(deletionHeight, errorID), errorID); err != nil {
		return err
	}

	// Store the actual sudo error
	err = k.Errors.Set(ctx, errorID, sudoErr)
	if err != nil {
		return err
	}

	types.EmitStoringErrorEvent(
		ctx,
		sudoErr,
		deletionHeight,
	)
	return nil
}

func (k Keeper) storeErrorCallback(ctx sdk.Context, sudoErr types.SudoError) error {
	errorID, err := k.ErrorID.Next(ctx)
	if err != nil {
		return err
	}

	k.SetSudoErrorCallback(ctx, errorID, sudoErr)
	return nil
}

// GetErrosByContractAddress returns all errors (in state) for a given contract address
func (k Keeper) GetErrorsByContractAddress(ctx sdk.Context, contractAddress []byte) (sudoErrs []types.SudoError, err error) {
	rng := collections.NewPrefixedPairRange[[]byte, uint64](contractAddress)
	err = k.ContractErrors.Walk(ctx, rng, func(key collections.Pair[[]byte, uint64], errorID uint64) (bool, error) {
		sudoErr, err := k.Errors.Get(ctx, errorID)
		if err != nil {
			return true, err
		}
		sudoErrs = append(sudoErrs, sudoErr)
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return sudoErrs, nil
}

// ExportErrors returns all errors in state. Used for genesis export
func (k Keeper) ExportErrors(ctx sdk.Context) (sudoErrs []types.SudoError, err error) {
	iter, err := k.Errors.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	sudoErrs, err = iter.Values()
	if err != nil {
		return nil, err
	}
	return sudoErrs, nil
}

// PruneErrorsCurrentBlock removes all errors that are queued to be deleted the given block height
func (k Keeper) PruneErrorsCurrentBlock(ctx sdk.Context) (err error) {
	var errorIDs []uint64
	height := ctx.BlockHeight()
	rng := collections.NewPrefixedPairRange[int64, uint64](height)
	err = k.DeletionBlocks.Walk(ctx, rng, func(key collections.Pair[int64, uint64], errorID uint64) (bool, error) {
		errorIDs = append(errorIDs, errorID)
		return false, nil
	})
	if err != nil {
		return err
	}
	for _, errorID := range errorIDs {
		sudoErr, err := k.Errors.Get(ctx, errorID)
		if err != nil {
			return err
		}
		// Removing the error data
		if err := k.Errors.Remove(ctx, errorID); err != nil {
			return err
		}
		// Removing the contract errors
		contractAddress := sdk.MustAccAddressFromBech32(sudoErr.ContractAddress)
		if err := k.ContractErrors.Remove(ctx, collections.Join(contractAddress.Bytes(), errorID)); err != nil {
			return err
		}
		// Removing the deletion block
		if err := k.DeletionBlocks.Remove(ctx, collections.Join(height, errorID)); err != nil {
			return err
		}
	}
	return nil
}

// SetSudoErrorCallback stores a sudo error callback in the transient store
func (k Keeper) SetSudoErrorCallback(ctx sdk.Context, errorId uint64, sudoErr types.SudoError) {
	tStore := ctx.TransientStore(k.tStoreKey)
	errToStore := k.cdc.MustMarshal(&sudoErr)
	tStore.Set(types.GetErrorsForSudoCallStoreKey(errorId), errToStore)
}

// GetAllSudoErrorCallbacks returns all sudo error callbacks from the transient store
func (k Keeper) GetAllSudoErrorCallbacks(ctx sdk.Context) (sudoErrs []types.SudoError) {
	tStore := ctx.TransientStore(k.tStoreKey)
	itr := sdk.KVStorePrefixIterator(tStore, types.ErrorsForSudoCallbackKey)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		var sudoErr types.SudoError
		k.cdc.MustUnmarshal(itr.Value(), &sudoErr)
		sudoErrs = append(sudoErrs, sudoErr)
	}
	return sudoErrs
}

// IterateSudoErrorCallbacks iterates over all sudo error callbacks from the transient store
func (k Keeper) IterateSudoErrorCallbacks(ctx sdk.Context, exec func(types.SudoError) bool) {
	tStore := ctx.TransientStore(k.tStoreKey)
	itr := sdk.KVStorePrefixIterator(tStore, types.ErrorsForSudoCallbackKey)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		var sudoErr types.SudoError
		k.cdc.MustUnmarshal(itr.Value(), &sudoErr)
		if exec(sudoErr) {
			break
		}
	}
}
