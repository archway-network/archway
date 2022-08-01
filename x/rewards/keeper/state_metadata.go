package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storeTypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/rewards/types"
)

// ContractMetadataState provides access to the types.ContractMetadata objects storage operations.
type ContractMetadataState struct {
	stateStore storeTypes.KVStore
	cdc        codec.Codec
	ctx        sdk.Context
}

// SetContractMetadata creates or modifies a types.ContractMetadata object.
func (s ContractMetadataState) SetContractMetadata(contractAddr sdk.AccAddress, obj types.ContractMetadata) {
	store := prefix.NewStore(s.stateStore, types.ContractMetadataPrefix)
	store.Set(
		s.buildContractMetadataKey(contractAddr),
		s.cdc.MustMarshal(&obj),
	)
}

// GetContractMetadata returns a types.ContractMetadata object by contract address.
func (s ContractMetadataState) GetContractMetadata(contractAddr sdk.AccAddress) (types.ContractMetadata, bool) {
	obj := s.getContractMetadata(contractAddr)
	if obj == nil {
		return types.ContractMetadata{}, false
	}

	return *obj, true
}

// Import initializes state from the module genesis data.
func (s ContractMetadataState) Import(objs []types.ContractMetadata) {
	for _, obj := range objs {
		s.SetContractMetadata(obj.MustGetContractAddress(), obj)
	}
}

// Export returns the module genesis data for the state.
func (s ContractMetadataState) Export() (objs []types.ContractMetadata) {
	store := prefix.NewStore(s.stateStore, types.ContractMetadataPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var obj types.ContractMetadata
		s.cdc.MustUnmarshal(iterator.Value(), &obj)

		objs = append(objs, obj)
	}

	return
}

// buildContractMetadataKey returns the key used to store a types.ContractMetadata object.
func (s ContractMetadataState) buildContractMetadataKey(contractAddr sdk.AccAddress) []byte {
	return contractAddr.Bytes()
}

// parseContractMetadataKey parses and validates types.ContractMetadata storage key.
func (s ContractMetadataState) parseContractMetadataKey(key []byte) sdk.AccAddress {
	addr := sdk.AccAddress(key)
	if err := sdk.VerifyAddressFormat(addr); err != nil {
		panic(fmt.Errorf("invalid contract address key: %w", err))
	}

	return addr
}

// getContractMetadata returns a types.ContractMetadata object by contract address.
func (s ContractMetadataState) getContractMetadata(contractAddr sdk.AccAddress) *types.ContractMetadata {
	store := prefix.NewStore(s.stateStore, types.ContractMetadataPrefix)

	bz := store.Get(s.buildContractMetadataKey(contractAddr))
	if bz == nil {
		return nil
	}

	var obj types.ContractMetadata
	s.cdc.MustUnmarshal(bz, &obj)

	return &obj
}
