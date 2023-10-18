package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storeTypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/rewards/types"
)

// BlockRewardsState provides access to the types.BlockRewards objects storage operations.
type BlockRewardsState struct {
	stateStore storeTypes.KVStore
	cdc        codec.Codec
	ctx        sdk.Context
}

// DeleteBlockRewards deletes a types.BlockRewards object.
func (s BlockRewardsState) DeleteBlockRewards(height int64) {
	store := prefix.NewStore(s.stateStore, types.BlockRewardsPrefix)
	store.Delete(s.buildBlockRewardsKey(height))
}

// buildBlockRewardsKey returns the key used to store a types.BlockRewards object.
func (s BlockRewardsState) buildBlockRewardsKey(height int64) []byte {
	return sdk.Uint64ToBigEndian(uint64(height))
}
