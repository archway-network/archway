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
		err := k.storeErrorCallback(ctx, contractAddr, sudoErr)
		if err != nil {
			return err
		}
		return nil
	}

	// for contracts which dont have an error subscription, store the error in state to be deleted after a set height
	return k.StoreErrorInState(ctx, contractAddr, sudoErr)
}

func (k Keeper) StoreErrorInState(ctx sdk.Context, contractAddr sdk.AccAddress, sudoErr types.SudoError) error {
	errorID, err := k.getNextErrorID(ctx)
	if err != nil {
		return err
	}

	// Store contract errors
	if err = k.ContractErrors.Set(ctx, collections.Join(contractAddr.Bytes(), errorID), errorID); err != nil {
		return err
	}

	// Store the deletion block
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	deletionHeight := ctx.BlockHeight() + params.ErrorStoredTime
	if err = k.DeletionBlocks.Set(ctx, collections.Join(deletionHeight, errorID), errorID); err != nil {
		return err
	}

	types.EmitStoringErrorEvent(
		ctx,
		sudoErr,
		deletionHeight,
	)
	// Store the error
	return k.Errors.Set(ctx, errorID, sudoErr)
}

func (k Keeper) storeErrorCallback(ctx sdk.Context, contractAddr sdk.AccAddress, sudoErr types.SudoError) error {
	errorID, err := k.getNextErrorID(ctx)
	if err != nil {
		return err
	}

	if k.HasSubscription(ctx, contractAddr) {
		k.SetSudoErrorCallback(ctx, errorID, sudoErr)
		return nil
	}
	return err
}

func (k Keeper) getNextErrorID(ctx sdk.Context) (int64, error) {
	errorID, err := k.GetErrorCount(ctx)
	if err != nil {
		return 0, err
	}
	errorID += 1
	if err = k.ErrorsCount.Set(ctx, errorID); err != nil {
		return 0, err
	}
	return errorID, nil
}

// GetErrosByContractAddress returns all errors by a given contract address
func (k Keeper) GetErrorsByContractAddress(ctx sdk.Context, contractAddress []byte) (sudoErrs []types.SudoError, err error) {
	rng := collections.NewPrefixedPairRange[[]byte, int64](contractAddress)
	err = k.ContractErrors.Walk(ctx, rng, func(key collections.Pair[[]byte, int64], errorID int64) (bool, error) {
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
	var errorIDs []int64
	height := ctx.BlockHeight()
	rng := collections.NewPrefixedPairRange[int64, int64](height)
	err = k.DeletionBlocks.Walk(ctx, rng, func(key collections.Pair[int64, int64], errorID int64) (bool, error) {
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

// GetErrorCount returns the total number of errors - used for generating errorID
func (k Keeper) GetErrorCount(ctx sdk.Context) (int64, error) {
	return k.ErrorsCount.Get(ctx)
}

// SetSudoErrorCallback stores a sudo error callback in the transient store
func (k Keeper) SetSudoErrorCallback(ctx sdk.Context, errorId int64, sudoErr types.SudoError) {
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
