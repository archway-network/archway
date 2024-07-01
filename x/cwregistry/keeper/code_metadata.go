package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/cwregistry/types"
)

// GetAllCodeMetadata returns all code metadata stored
func (k Keeper) GetAllCodeMetadata(ctx sdk.Context) (codeMetadata []types.CodeMetadata, err error) {
	err = k.CodeMetadata.Walk(ctx, nil, func(key uint64, value types.CodeMetadata) (stop bool, err error) {
		codeMetadata = append(codeMetadata, value)
		return false, nil
	})
	return codeMetadata, err
}

// GetCodeMetadata returns the code metadata for the given codeID
func (k Keeper) GetCodeMetadata(ctx sdk.Context, codeID uint64) (types.CodeMetadata, error) {
	return k.CodeMetadata.Get(ctx, codeID)
}

// HasCodeMetadata returns true if the code metadata for the given codeID exists
func (k Keeper) HasCodeMetadata(ctx sdk.Context, codeID uint64) bool {
	has, err := k.CodeMetadata.Has(ctx, codeID)
	if err != nil {
		return false
	}
	return has
}

// SetCodeMetadata sets the metadata for the code with the given codeID
func (k Keeper) SetCodeMetadata(ctx sdk.Context, sender sdk.AccAddress, codeID uint64, codeMetadata types.CodeMetadata) error {
	codeInfo := k.wasmKeeper.GetCodeInfo(ctx, codeID)
	if codeInfo == nil {
		return types.ErrNoSuchCode
	}
	if codeInfo.Creator != sender.String() {
		return types.ErrUnauthorized
	}
	codeMetadata.CodeId = codeID
	return k.saveCodeMetadata(ctx, codeMetadata)
}

// UnsafeSetCodeMetadata sets the metadata for the code with the given codeID without checking permissions
// Should only be used in genesis
func (k Keeper) UnsafeSetCodeMetadata(ctx sdk.Context, codeMetadata types.CodeMetadata) error {
	codeInfo := k.wasmKeeper.GetCodeInfo(ctx, codeMetadata.CodeId)
	if codeInfo == nil {
		return types.ErrNoSuchCode
	}
	return k.saveCodeMetadata(ctx, codeMetadata)
}

// saveCodeMetadata saves the code metadata to the store
func (k Keeper) saveCodeMetadata(ctx sdk.Context, codeMetadata types.CodeMetadata) error {
	if err := codeMetadata.Validate(); err != nil {
		return err
	}
	return k.CodeMetadata.Set(ctx, codeMetadata.CodeId, codeMetadata)
}
