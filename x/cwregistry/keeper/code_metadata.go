package keeper

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/archway-network/archway/x/cwregistry/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetAllCallbacks returns all code metadata stored
func (k Keeper) GetAllCallbacks(ctx sdk.Context) (codeMetadata []types.CodeMetadata, err error) {
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

// SetContractMetadata sets the metadata for the contract with the given address
func (k Keeper) SetContractMetadata(ctx sdk.Context, sender sdk.AccAddress, contractAddress sdk.AccAddress, codeMetadata types.CodeMetadata) error {
	contractInfo := k.wasmKeeper.GetContractInfo(ctx, contractAddress)
	if contractInfo == nil {
		return types.ErrNoSuchContract
	}
	if contractInfo.Creator != sender.String() && contractInfo.Admin != sender.String() && sender.String() != contractAddress.String() {
		return types.ErrUnauthorized
	}
	codeID := contractInfo.CodeID
	codeMetadata.CodeId = codeID
	return k.saveCodeMetadata(ctx, codeMetadata)
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

// saveCodeMetadata saves the code metadata to the store
func (k Keeper) saveCodeMetadata(ctx sdk.Context, codeMetadata types.CodeMetadata) error {
	schemaContent := codeMetadata.Schema
	if len(schemaContent) > 255 { // we dont want to store large schemas in the store
		// todo: save to fs
		schemaHash := hash(schemaContent)
		codeMetadata.Schema = schemaHash
	}
	if err := codeMetadata.Validate(); err != nil {
		return err
	}
	return k.CodeMetadata.Set(ctx, codeMetadata.CodeId, codeMetadata)
}

// hash returns the sha256 hash of the given schema
func hash(schemaContent string) string {
	hasher := sha256.New()
	hasher.Write([]byte(schemaContent))
	return hex.EncodeToString(hasher.Sum(nil))
}
