package keeper

import (
	"fmt"
	"math"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storeTypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/rewards/types"
)

// TxRewardsState provides access to the types.TxRewards objects storage operations.
type TxRewardsState struct {
	stateStore storeTypes.KVStore
	cdc        codec.Codec
	ctx        sdk.Context
}

// CreateTxRewards creates a new types.TxRewards object.
func (s TxRewardsState) CreateTxRewards(txID uint64, height int64, rewards sdk.Coins) types.TxRewards {
	obj := types.TxRewards{
		TxId:       txID,
		Height:     height,
		FeeRewards: rewards,
	}

	s.setTxRewards(&obj)
	s.setBlockIndex(obj.Height, obj.TxId)

	return obj
}

// GetTxRewards returns a types.TxRewards object by txID.
func (s TxRewardsState) GetTxRewards(txID uint64) (types.TxRewards, bool) {
	obj := s.getTxRewards(txID)
	if obj == nil {
		return types.TxRewards{}, false
	}

	return *obj, true
}

// GetTxRewardsByBlock returns a list of types.TxRewards objects by block height.
func (s TxRewardsState) GetTxRewardsByBlock(height int64) (objs []types.TxRewards) {
	store := prefix.NewStore(s.stateStore, types.TxRewardsBlockIndexPrefix)

	iterator := sdk.KVStorePrefixIterator(store, s.buildBlockIndexPrefix(height))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		_, txID := s.parseBlockIndexKey(iterator.Key())

		obj, found := s.GetTxRewards(txID)
		if !found {
			panic(fmt.Errorf("invalid TxRewards Block index state: txId (%d): not found", txID))
		}
		objs = append(objs, obj)
	}

	return
}

// Import initializes state from the module genesis data.
func (s TxRewardsState) Import(objs []types.TxRewards) {
	for _, obj := range objs {
		s.setTxRewards(&obj)
		s.setBlockIndex(obj.Height, obj.TxId)
	}
}

// Export returns the module genesis data for the state.
func (s TxRewardsState) Export() (objs []types.TxRewards) {
	store := prefix.NewStore(s.stateStore, types.TxRewardsPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var obj types.TxRewards
		s.cdc.MustUnmarshal(iterator.Value(), &obj)
		objs = append(objs, obj)
	}

	return
}

// deleteTxRewardsByBlock deletes types.TxRewards objects by block height cleaning up the block index.
// Returns the list of deleted txIDs.
func (s TxRewardsState) deleteTxRewardsByBlock(height int64) []uint64 {
	store := prefix.NewStore(s.stateStore, types.TxRewardsBlockIndexPrefix)

	iterator := sdk.KVStorePrefixIterator(store, s.buildBlockIndexPrefix(height))
	defer iterator.Close()

	var blockIndexKeys [][]byte
	var removedTxIDs []uint64
	for ; iterator.Valid(); iterator.Next() {
		_, txID := s.parseBlockIndexKey(iterator.Key())
		s.deleteTxRewards(txID)

		removedTxIDs = append(removedTxIDs, txID)
		blockIndexKeys = append(blockIndexKeys, iterator.Key())
	}
	for _, key := range blockIndexKeys {
		store.Delete(key)
	}

	return removedTxIDs
}

// buildTxRewardsKey returns the key used to store a types.TxRewards object.
func (s TxRewardsState) buildTxRewardsKey(txID uint64) []byte {
	return sdk.Uint64ToBigEndian(txID)
}

// setTxRewards sets a types.TxRewards object.
func (s TxRewardsState) setTxRewards(obj *types.TxRewards) {
	store := prefix.NewStore(s.stateStore, types.TxRewardsPrefix)
	store.Set(
		s.buildTxRewardsKey(obj.TxId),
		s.cdc.MustMarshal(obj),
	)
}

// getContractOpInfo returns a types.ContractOperationInfo object by ID.
func (s TxRewardsState) getTxRewards(txID uint64) *types.TxRewards {
	store := prefix.NewStore(s.stateStore, types.TxRewardsPrefix)

	bz := store.Get(s.buildTxRewardsKey(txID))
	if bz == nil {
		return nil
	}

	var obj types.TxRewards
	s.cdc.MustUnmarshal(bz, &obj)

	return &obj
}

// deleteTxRewards deletes a types.TxRewards object.
func (s TxRewardsState) deleteTxRewards(txID uint64) {
	store := prefix.NewStore(s.stateStore, types.TxRewardsPrefix)
	store.Delete(s.buildTxRewardsKey(txID))
}

// buildBlockIndexPrefix returns the key prefix used to maintain types.TxRewards's block index.
func (s TxRewardsState) buildBlockIndexPrefix(height int64) []byte {
	return sdk.Uint64ToBigEndian(uint64(height))
}

// buildBlockIndexKey returns the key used to maintain types.TxRewards's block index.
func (s TxRewardsState) buildBlockIndexKey(height int64, txID uint64) []byte {
	return append(
		s.buildBlockIndexPrefix(height),
		sdk.Uint64ToBigEndian(txID)...,
	)
}

// parseBlockIndexKey parses the types.TxRewards's block index key.
func (s TxRewardsState) parseBlockIndexKey(key []byte) (height int64, txID uint64) {
	if len(key) != 16 {
		panic(fmt.Errorf("invalid TxRewards Block index key length: %d", len(key)))
	}

	heightRaw := sdk.BigEndianToUint64(key[:8])
	if heightRaw > math.MaxInt64 {
		panic(fmt.Errorf("invalid TxRewards Block index key height: %d", heightRaw))
	}
	height = int64(heightRaw)

	txID = sdk.BigEndianToUint64(key[8:])

	return
}

// setBlockIndex adds the types.TxRewards's block index entry.
func (s TxRewardsState) setBlockIndex(height int64, txID uint64) {
	store := prefix.NewStore(s.stateStore, types.TxRewardsBlockIndexPrefix)
	store.Set(
		s.buildBlockIndexKey(height, txID),
		[]byte{},
	)
}
