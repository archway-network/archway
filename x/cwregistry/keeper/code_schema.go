package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

func (k Keeper) GetSchema(ctx sdk.Context, codeID uint64) (string, error) {
	return "", nil
}

func (k Keeper) SetSchema(ctx sdk.Context, codeID uint64, schema string) error {
	return nil
}

func (k Keeper) HasSchema(ctx sdk.Context, codeID uint64) bool {
	return false
}
