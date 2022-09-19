package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storeTypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/tracking/types"
)

// TxInfoState provides access to the types.TxInfo objects storage operations.
type TxInfoState struct {
	stateStore storeTypes.KVStore
	cdc        codec.Codec
	ctx        sdk.Context
}

// GetCurrentTxID returns the latest types.TxInfo unique ID.
func (s TxInfoState) GetCurrentTxID() uint64 {
	lastIDBz := s.stateStore.Get(types.TxInfoIDKey)
	lastID := sdk.BigEndianToUint64(lastIDBz) // returns 0 if nil

	return lastID
}

// CreateEmptyTxInfo creates a new types.TxInfo object with unique ID.
func (s TxInfoState) CreateEmptyTxInfo() types.TxInfo {
	obj := types.TxInfo{
		Id:     s.nextID(),
		Height: s.ctx.BlockHeight(),
	}

	s.SetTxInfo(obj)
	s.setBlockIndex(s.ctx.BlockHeight(), obj.Id)
	s.setLastID(obj.Id)

	return obj
}

// SetTxInfo sets a types.TxInfo object overwriting an existing one.
// CONTRACT: Block index is not updated.
func (s TxInfoState) SetTxInfo(obj types.TxInfo) {
	store := prefix.NewStore(s.stateStore, types.TxInfoPrefix)
	store.Set(
		s.buildTxInfoKey(obj.Id),
		s.cdc.MustMarshal(&obj),
	)
}

// GetTxInfo returns a types.TxInfo object by ID.
func (s TxInfoState) GetTxInfo(id uint64) (types.TxInfo, bool) {
	obj := s.getTxInfo(id)
	if obj == nil {
		return types.TxInfo{}, false
	}

	return *obj, true
}

// GetTxInfosByBlock returns a list of types.TxInfo objects by block height.
func (s TxInfoState) GetTxInfosByBlock(height int64) (objs []types.TxInfo) {
	store := prefix.NewStore(s.stateStore, types.TxInfoBlockIndexPrefix)

	iterator := sdk.KVStorePrefixIterator(store, s.buildBlockIndexPrefix(height))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		_, id := s.parseBlockIndexKey(iterator.Key())

		obj, found := s.GetTxInfo(id)
		if !found {
			panic(fmt.Errorf("invalid TxInfo Block index state: id (%d): not found", id))
		}
		objs = append(objs, obj)
	}

	return
}

// DeleteTxInfosByBlock deletes all types.TxInfo objects by block height clearing the block index.
// Returns the list of deleted IDs.
func (s TxInfoState) DeleteTxInfosByBlock(height int64) []uint64 {
	store := prefix.NewStore(s.stateStore, types.TxInfoBlockIndexPrefix)

	iterator := sdk.KVStorePrefixIterator(store, s.buildBlockIndexPrefix(height))
	defer iterator.Close()

	var blockIndexKeys [][]byte
	var removedIDs []uint64
	for ; iterator.Valid(); iterator.Next() {
		_, id := s.parseBlockIndexKey(iterator.Key())
		s.deleteTxInfo(id)

		removedIDs = append(removedIDs, id)
		blockIndexKeys = append(blockIndexKeys, iterator.Key())
	}
	for _, key := range blockIndexKeys {
		store.Delete(key)
	}

	return removedIDs
}

// Import initializes state from the module genesis data.
func (s TxInfoState) Import(lastID uint64, objs []types.TxInfo) {
	for _, obj := range objs {
		s.SetTxInfo(obj)
		s.setBlockIndex(obj.Height, obj.Id)
	}
	s.setLastID(lastID)
}

// Export returns the module genesis data for the state.
func (s TxInfoState) Export() (lastID uint64, objs []types.TxInfo) {
	store := prefix.NewStore(s.stateStore, types.TxInfoPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var obj types.TxInfo
		s.cdc.MustUnmarshal(iterator.Value(), &obj)
		objs = append(objs, obj)
	}
	lastID = s.nextID() - 1

	return
}

// setLastID sets the last types.TxInfo unique ID.
func (s TxInfoState) setLastID(id uint64) {
	s.stateStore.Set(
		types.TxInfoIDKey,
		sdk.Uint64ToBigEndian(id),
	)
}

// nextID returns the next types.TxInfo unique ID.
func (s TxInfoState) nextID() uint64 {
	lastIDBz := s.stateStore.Get(types.TxInfoIDKey)
	lastID := sdk.BigEndianToUint64(lastIDBz) // returns 0 if nil

	return lastID + 1
}

// buildTxInfoKey returns the key used to store a types.TxInfo object.
func (s TxInfoState) buildTxInfoKey(id uint64) []byte {
	return sdk.Uint64ToBigEndian(id)
}

// getTxInfo returns a types.TxInfo object by ID.
func (s TxInfoState) getTxInfo(id uint64) *types.TxInfo {
	store := prefix.NewStore(s.stateStore, types.TxInfoPrefix)

	bz := store.Get(s.buildTxInfoKey(id))
	if bz == nil {
		return nil
	}

	var obj types.TxInfo
	s.cdc.MustUnmarshal(bz, &obj)

	return &obj
}

// deleteTxInfo deletes a types.TxInfo object by ID.
func (s TxInfoState) deleteTxInfo(id uint64) {
	store := prefix.NewStore(s.stateStore, types.TxInfoPrefix)
	store.Delete(s.buildTxInfoKey(id))
}

// buildBlockIndexPrefix returns the key prefix used to maintain types.TxInfo's block index.
func (s TxInfoState) buildBlockIndexPrefix(height int64) []byte {
	return sdk.Uint64ToBigEndian(uint64(height))
}

// buildBlockIndexKey returns the key used to maintain types.TxInfo's block index.
func (s TxInfoState) buildBlockIndexKey(height int64, id uint64) []byte {
	return append(
		s.buildBlockIndexPrefix(height),
		sdk.Uint64ToBigEndian(id)...,
	)
}

// parseBlockIndexKey parses the types.TxInfo's block index key.
func (s TxInfoState) parseBlockIndexKey(key []byte) (height int64, id uint64) {
	if len(key) != 16 {
		panic(fmt.Errorf("invalid TxInfo Block index key length: %d", len(key)))
	}

	height = int64(sdk.BigEndianToUint64(key[:8]))
	id = sdk.BigEndianToUint64(key[8:])

	return
}

// setBlockIndex adds the types.TxInfo's block index entry.
func (s TxInfoState) setBlockIndex(height int64, id uint64) {
	store := prefix.NewStore(s.stateStore, types.TxInfoBlockIndexPrefix)
	store.Set(
		s.buildBlockIndexKey(height, id),
		[]byte{},
	)
}
