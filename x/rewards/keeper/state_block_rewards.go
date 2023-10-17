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

// GetBlockRewards returns a types.BlockRewards object by block height.
func (s BlockRewardsState) GetBlockRewards(height int64) (types.BlockRewards, bool) {
	store := prefix.NewStore(s.stateStore, types.BlockRewardsPrefix)
	key := s.buildBlockRewardsKey(height)

	bz := store.Get(key)
	if bz == nil {
		return types.BlockRewards{}, false
	}

	var obj types.BlockRewards
	s.cdc.MustUnmarshal(bz, &obj)

	return obj, true
}

// DeleteBlockRewards deletes a types.BlockRewards object.
func (s BlockRewardsState) DeleteBlockRewards(height int64) {
	store := prefix.NewStore(s.stateStore, types.BlockRewardsPrefix)
	store.Delete(s.buildBlockRewardsKey(height))
}

// Import initializes state from the module genesis data.
func (s BlockRewardsState) Import(objs []types.BlockRewards) {
	for _, obj := range objs {
		s.setBlockRewards(&obj)
	}
}

// buildBlockRewardsKey returns the key used to store a types.BlockRewards object.
func (s BlockRewardsState) buildBlockRewardsKey(height int64) []byte {
	return sdk.Uint64ToBigEndian(uint64(height))
}

// setBlockRewards sets a types.BlockRewards object.
func (s BlockRewardsState) setBlockRewards(obj *types.BlockRewards) {
	store := prefix.NewStore(s.stateStore, types.BlockRewardsPrefix)
	store.Set(
		s.buildBlockRewardsKey(obj.Height),
		s.cdc.MustMarshal(obj),
	)
}
