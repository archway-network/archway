package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storeTypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/archway-network/archway/x/rewards/types"
)

// RewardsRecordState provides access to the types.RewardsRecord objects storage operations.
type RewardsRecordState struct {
	stateStore storeTypes.KVStore
	cdc        codec.Codec
	ctx        sdk.Context
}

// CreateRewardsRecord creates a new types.RewardsRecord object with unique ID.
func (s RewardsRecordState) CreateRewardsRecord(rewardsAddr sdk.AccAddress, rewards sdk.Coins, calculatedHeight int64, calculatedTime time.Time) types.RewardsRecord {
	obj := types.RewardsRecord{
		Id:               s.getNextID(),
		RewardsAddress:   rewardsAddr.String(),
		Rewards:          rewards,
		CalculatedHeight: calculatedHeight,
		CalculatedTime:   calculatedTime,
	}

	s.setRewardsRecord(&obj)
	s.setAddressIndex(obj.Id, rewardsAddr)
	s.setLastID(obj.Id)

	return obj
}

// GetRewardsRecord returns a types.RewardsRecord object by ID.
func (s RewardsRecordState) GetRewardsRecord(id uint64) (types.RewardsRecord, bool) {
	store := prefix.NewStore(s.stateStore, types.RewardsRecordPrefix)

	bz := store.Get(s.buildRewardsRecordKey(id))
	if bz == nil {
		return types.RewardsRecord{}, false
	}

	var obj types.RewardsRecord
	s.cdc.MustUnmarshal(bz, &obj)

	return obj, true
}

// GetRewardsRecordByRewardsAddress returns a list of types.RewardsRecord objects by rewardsAddress.
func (s RewardsRecordState) GetRewardsRecordByRewardsAddress(rewardsAddr sdk.AccAddress) (objs []types.RewardsRecord) {
	store := prefix.NewStore(s.stateStore, types.RewardsRecordAddressIndexPrefix)

	iterator := sdk.KVStorePrefixIterator(store, s.buildAddressIndexPrefix(rewardsAddr))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		_, id := s.parseAddressIndexKey(iterator.Key())

		obj, found := s.GetRewardsRecord(id)
		if !found {
			panic(fmt.Errorf("invalid RewardsRecord RewardsAddress index state: id (%d): not found", id))
		}
		objs = append(objs, obj)
	}

	return
}

// GetRewardsRecordByRewardsAddressPaginated returns a list of types.RewardsRecord objects by rewardsAddress paginated.
func (s RewardsRecordState) GetRewardsRecordByRewardsAddressPaginated(rewardsAddr sdk.AccAddress, pageReq *query.PageRequest) ([]types.RewardsRecord, *query.PageResponse, error) {
	store := prefix.NewStore(
		prefix.NewStore(s.stateStore, types.RewardsRecordAddressIndexPrefix),
		s.buildAddressIndexPrefix(rewardsAddr),
	)

	var objs []types.RewardsRecord
	pageRes, err := query.Paginate(store, pageReq, func(key, _ []byte) error {
		id := s.parseIdKey(key)

		obj, found := s.GetRewardsRecord(id)
		if !found {
			panic(fmt.Errorf("invalid RewardsRecord RewardsAddress index state: id (%d): not found", id))
		}
		objs = append(objs, obj)

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return objs, pageRes, nil
}

// DeleteRewardsRecords deletes a list of types.RewardsRecord objects updating indexes.
func (s RewardsRecordState) DeleteRewardsRecords(objs ...types.RewardsRecord) {
	for _, obj := range objs {
		s.deleteRewardsRecord(obj.Id)
		s.deleteAddressIndexEntry(obj.Id, obj.MustGetRewardsAddress())
	}
}

// Import initializes state from the module genesis data.
func (s RewardsRecordState) Import(lastID uint64, objs []types.RewardsRecord) {
	for _, obj := range objs {
		s.setRewardsRecord(&obj)
		s.setAddressIndex(obj.Id, obj.MustGetRewardsAddress())
	}
	s.setLastID(lastID)
}

// Export returns the module genesis data for the state.
func (s RewardsRecordState) Export() (lastID uint64, objs []types.RewardsRecord) {
	store := prefix.NewStore(s.stateStore, types.RewardsRecordPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var obj types.RewardsRecord
		s.cdc.MustUnmarshal(iterator.Value(), &obj)
		objs = append(objs, obj)
	}
	lastID = s.getNextID() - 1

	return
}

// setLastID sets the last types.RewardsRecord unique ID.
func (s RewardsRecordState) setLastID(id uint64) {
	s.stateStore.Set(
		types.RewardsRecordIDKey,
		sdk.Uint64ToBigEndian(id),
	)
}

// getNextID returns the next types.RewardsRecord unique ID.
func (s RewardsRecordState) getNextID() uint64 {
	lastIDBz := s.stateStore.Get(types.RewardsRecordIDKey)
	lastID := sdk.BigEndianToUint64(lastIDBz) // returns 0 if nil

	return lastID + 1
}

// buildRewardsRecordKey returns the key used to store a types.RewardsRecord object.
func (s RewardsRecordState) buildRewardsRecordKey(id uint64) []byte {
	return sdk.Uint64ToBigEndian(id)
}

// setRewardsRecord sets a types.RewardsRecord object.
func (s RewardsRecordState) setRewardsRecord(obj *types.RewardsRecord) {
	store := prefix.NewStore(s.stateStore, types.RewardsRecordPrefix)
	store.Set(
		s.buildRewardsRecordKey(obj.Id),
		s.cdc.MustMarshal(obj),
	)
}

// deleteRewardsRecord deletes a types.RewardsRecord object.
func (s RewardsRecordState) deleteRewardsRecord(id uint64) {
	store := prefix.NewStore(s.stateStore, types.RewardsRecordPrefix)
	store.Delete(s.buildRewardsRecordKey(id))
}

// buildAddressIndexPrefix returns the key prefix used to maintain types.RewardsRecord's RewardsAddress index.
func (s RewardsRecordState) buildAddressIndexPrefix(rewardsAddr sdk.AccAddress) []byte {
	return address.MustLengthPrefix(rewardsAddr)
}

// buildAddressIndexKey returns the key used to maintain types.RewardsRecord's RewardsAddress index.
func (s RewardsRecordState) buildAddressIndexKey(id uint64, rewardsAddr sdk.AccAddress) []byte {
	return append(
		s.buildAddressIndexPrefix(rewardsAddr),
		sdk.Uint64ToBigEndian(id)...,
	)
}

// parseAddressIndexKey parses the types.RewardsRecord's RewardsAddress index key.
func (s RewardsRecordState) parseAddressIndexKey(key []byte) (rewardsAddr sdk.AccAddress, id uint64) {
	// Check min length: 1 length prefixed addr + 8 uint64
	if len(key) <= 9 {
		panic(fmt.Errorf("invalid RewardsRecord RewardsAddress index key min length: %d", len(key)))
	}

	// Check key length
	addrLen := int(key[0])
	if len(key) != 1+addrLen+8 {
		panic(fmt.Errorf("invalid RewardsRecord RewardsAddress index key length: %d", len(key)))
	}

	// Parse keys
	rewardsAddr = sdk.AccAddress(key[1 : 1+addrLen])
	if err := sdk.VerifyAddressFormat(rewardsAddr); err != nil {
		panic(fmt.Errorf("invalid RewardsRecord RewardsAddress index key (address): %s", err))
	}

	id = sdk.BigEndianToUint64(key[1+addrLen:])

	return
}

// parseIdKey parses the 2nd part of the types.RewardsRecord's RewardsAddress index key (ID).
func (s RewardsRecordState) parseIdKey(key []byte) uint64 {
	// Check min length
	if len(key) != 8 {
		panic(fmt.Errorf("invalid RewardsRecord RewardsAddress index key min length (ID): %d", len(key)))
	}

	id := sdk.BigEndianToUint64(key)

	return id
}

// setAddressIndex adds the types.RewardsRecord's RewardsAddress index entry.
func (s RewardsRecordState) setAddressIndex(id uint64, rewardsAddr sdk.AccAddress) {
	store := prefix.NewStore(s.stateStore, types.RewardsRecordAddressIndexPrefix)
	store.Set(
		s.buildAddressIndexKey(id, rewardsAddr),
		[]byte{},
	)
}

// deleteAddressIndexEntry deletes the types.RewardsRecord's RewardsAddress index entry.
func (s RewardsRecordState) deleteAddressIndexEntry(id uint64, rewardsAddr sdk.AccAddress) {
	store := prefix.NewStore(s.stateStore, types.RewardsRecordAddressIndexPrefix)
	store.Delete(s.buildAddressIndexKey(id, rewardsAddr))
}
