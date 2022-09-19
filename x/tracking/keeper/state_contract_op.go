package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storeTypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/tracking/types"
)

// ContractOpInfoState provides access to the types.ContractOperationInfo objects storage operations.
type ContractOpInfoState struct {
	stateStore storeTypes.KVStore
	cdc        codec.Codec
	ctx        sdk.Context
}

// CreateContractOpInfo creates a new types.ContractOperationInfo object with unique ID.
func (s ContractOpInfoState) CreateContractOpInfo(txID uint64, contractAddr sdk.AccAddress, opType types.ContractOperation, vmGas, sdkGas uint64) types.ContractOperationInfo {
	obj := types.ContractOperationInfo{
		Id:              s.getNextID(),
		TxId:            txID,
		ContractAddress: contractAddr.String(),
		OperationType:   opType,
		VmGas:           vmGas,
		SdkGas:          sdkGas,
	}

	s.setContractOpInfo(&obj)
	s.setTxIndex(obj.TxId, obj.Id)
	s.setLastID(obj.Id)

	return obj
}

// GetContractOpInfo returns a types.ContractOperationInfo object by ID.
func (s ContractOpInfoState) GetContractOpInfo(id uint64) (types.ContractOperationInfo, bool) {
	obj := s.getContractOpInfo(id)
	if obj == nil {
		return types.ContractOperationInfo{}, false
	}

	return *obj, true
}

// GetContractOpInfoByTxID returns a list of types.ContractOperationInfo objects by tx ID.
func (s ContractOpInfoState) GetContractOpInfoByTxID(txID uint64) (objs []types.ContractOperationInfo) {
	store := prefix.NewStore(s.stateStore, types.ContractOpInfoTxIndexPrefix)

	iterator := sdk.KVStorePrefixIterator(store, s.buildTxIndexPrefix(txID))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		_, id := s.parseTxIndexKey(iterator.Key())

		obj, found := s.GetContractOpInfo(id)
		if !found {
			panic(fmt.Errorf("invalid ContractOpInfo TxInfo index state: id (%d): not found", id))
		}
		objs = append(objs, obj)
	}

	return
}

// DeleteContractOpsByTxID deletes all types.ContractOperationInfo objects by tx ID.
// Returns the list of deleted IDs.
func (s ContractOpInfoState) DeleteContractOpsByTxID(txID uint64) []uint64 {
	store := prefix.NewStore(s.stateStore, types.ContractOpInfoTxIndexPrefix)

	iterator := sdk.KVStorePrefixIterator(store, s.buildTxIndexPrefix(txID))
	defer iterator.Close()

	var txIndexKeys [][]byte
	var removedIDs []uint64
	for ; iterator.Valid(); iterator.Next() {
		_, id := s.parseTxIndexKey(iterator.Key())
		s.deleteContractOpInfo(id)

		removedIDs = append(removedIDs, id)
		txIndexKeys = append(txIndexKeys, iterator.Key())
	}
	for _, key := range txIndexKeys {
		store.Delete(key)
	}

	return removedIDs
}

// Import initializes state from the module genesis data.
func (s ContractOpInfoState) Import(lastID uint64, objs []types.ContractOperationInfo) {
	for _, obj := range objs {
		s.setContractOpInfo(&obj)
		s.setTxIndex(obj.TxId, obj.Id)
	}
	s.setLastID(lastID)
}

// Export returns the module genesis data for the state.
func (s ContractOpInfoState) Export() (lastID uint64, objs []types.ContractOperationInfo) {
	store := prefix.NewStore(s.stateStore, types.ContractOpInfoPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var obj types.ContractOperationInfo
		s.cdc.MustUnmarshal(iterator.Value(), &obj)
		objs = append(objs, obj)
	}
	lastID = s.getNextID() - 1

	return
}

// setLastID sets the last types.ContractOperationInfo unique ID.
func (s ContractOpInfoState) setLastID(id uint64) {
	s.stateStore.Set(
		types.ContractOpInfoIDKey,
		sdk.Uint64ToBigEndian(id),
	)
}

// getNextID returns the next types.ContractOperationInfo unique ID.
func (s ContractOpInfoState) getNextID() uint64 {
	lastIDBz := s.stateStore.Get(types.ContractOpInfoIDKey)
	lastID := sdk.BigEndianToUint64(lastIDBz) // returns 0 if nil

	return lastID + 1
}

// buildTxInfoKey returns the key used to store a types.ContractOperationInfo object.
func (s ContractOpInfoState) buildContractOpInfoKey(id uint64) []byte {
	return sdk.Uint64ToBigEndian(id)
}

// setContractOpInfo sets a types.ContractOperationInfo object.
func (s ContractOpInfoState) setContractOpInfo(obj *types.ContractOperationInfo) {
	store := prefix.NewStore(s.stateStore, types.ContractOpInfoPrefix)
	store.Set(
		s.buildContractOpInfoKey(obj.Id),
		s.cdc.MustMarshal(obj),
	)
}

// getContractOpInfo returns a types.ContractOperationInfo object by ID.
func (s ContractOpInfoState) getContractOpInfo(id uint64) *types.ContractOperationInfo {
	store := prefix.NewStore(s.stateStore, types.ContractOpInfoPrefix)

	bz := store.Get(s.buildContractOpInfoKey(id))
	if bz == nil {
		return nil
	}

	var obj types.ContractOperationInfo
	s.cdc.MustUnmarshal(bz, &obj)

	return &obj
}

// deleteContractOpInfo deletes a types.ContractOperationInfo object by ID.
func (s ContractOpInfoState) deleteContractOpInfo(id uint64) {
	store := prefix.NewStore(s.stateStore, types.ContractOpInfoPrefix)
	store.Delete(s.buildContractOpInfoKey(id))
}

// buildTxIndexPrefix returns the key prefix used to maintain types.ContractOperationInfo's tx index.
func (s ContractOpInfoState) buildTxIndexPrefix(txID uint64) []byte {
	return sdk.Uint64ToBigEndian(txID)
}

// buildTxIndexKey returns the key used to maintain types.ContractOperationInfo's tx index.
func (s ContractOpInfoState) buildTxIndexKey(txID, id uint64) []byte {
	return append(
		s.buildTxIndexPrefix(txID),
		sdk.Uint64ToBigEndian(id)...,
	)
}

// parseTxIndexKey parses the types.ContractOperationInfo's tx index key.
func (s ContractOpInfoState) parseTxIndexKey(key []byte) (txID, id uint64) {
	if len(key) != 16 {
		panic(fmt.Errorf("invalid ContractOpInfo TxInfo index key length: %d", len(key)))
	}

	txID = sdk.BigEndianToUint64(key[:8])
	id = sdk.BigEndianToUint64(key[8:])

	return
}

// setTxIndex adds the types.ContractOperationInfo's tx index entry.
func (s ContractOpInfoState) setTxIndex(txID, id uint64) {
	store := prefix.NewStore(s.stateStore, types.ContractOpInfoTxIndexPrefix)
	store.Set(
		s.buildTxIndexKey(txID, id),
		[]byte{},
	)
}
