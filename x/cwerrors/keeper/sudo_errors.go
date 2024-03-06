package keeper

import (
	"cosmossdk.io/collections"
	"github.com/archway-network/archway/x/cwerrors/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	// Get the error id
	errorID, err := k.GetErrorCount(ctx)
	if err != nil {
		return err
	}
	errorID += 1
	if err = k.ErrorsCount.Set(ctx, errorID); err != nil {
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
	deletionHeight := ctx.BlockHeight() + params.GetErrorStoredTime()
	if err = k.DeletionBlocks.Set(ctx, collections.Join(deletionHeight, errorID), errorID); err != nil {
		return err
	}

	// Store the error
	return k.Errors.Set(ctx, errorID, sudoErr)
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
