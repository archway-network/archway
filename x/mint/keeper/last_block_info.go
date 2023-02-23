package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/mint/types"
)

// SetLastBlockInfo stores the last block info. Fails and returns an error if the input is not valid
func (k Keeper) SetLastBlockInfo(ctx sdk.Context, lbi types.LastBlockInfo) error {
	if err := lbi.Validate(); err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	value := k.cdc.MustMarshal(&lbi)

	store.Set(types.LastBlockInfoPrefix, value)
	return nil
}

// GetLastBlockInfo fetches the latest block info
func (k Keeper) GetLastBlockInfo(ctx sdk.Context) (bool, types.LastBlockInfo) {
	store := ctx.KVStore(k.storeKey)

	var lbi types.LastBlockInfo
	bz := store.Get(types.LastBlockInfoPrefix)
	if bz == nil {
		return false, lbi
	}

	k.cdc.MustUnmarshal(bz, &lbi)
	return true, lbi
}
