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

// CreateBlockRewards creates a types.BlockRewards object.
func (s BlockRewardsState) CreateBlockRewards(height int64, rewards sdk.Coin, blockMaxGas uint64) types.BlockRewards {
	obj := types.BlockRewards{
		Height:           height,
		InflationRewards: rewards,
		MaxGas:           blockMaxGas,
	}
	s.setBlockRewards(&obj)

	return obj
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

// Export returns the module genesis data for the state.
func (s BlockRewardsState) Export() (objs []types.BlockRewards) {
	store := prefix.NewStore(s.stateStore, types.BlockRewardsPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var obj types.BlockRewards
		s.cdc.MustUnmarshal(iterator.Value(), &obj)

		objs = append(objs, obj)
	}

	return
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
