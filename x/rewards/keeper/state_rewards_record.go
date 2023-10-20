package keeper

import (
	"fmt"

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

// buildRewardsRecordKey returns the key used to store a types.RewardsRecord object.
func (s RewardsRecordState) buildRewardsRecordKey(id uint64) []byte {
	return sdk.Uint64ToBigEndian(id)
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
