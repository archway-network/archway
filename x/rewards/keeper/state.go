package keeper

import (
	"github.com/archway-network/archway/x/rewards/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// State is a wrapper around the module storage state.
type State struct {
	key sdk.StoreKey
	cdc codec.Codec
}

// NewState creates a new State instance.
func NewState(cdc codec.Codec, key sdk.StoreKey) State {
	return State{
		key: key,
		cdc: cdc,
	}
}

// ContractMetadataState returns types.ContractMetadata repository.
func (s State) ContractMetadataState(ctx sdk.Context) ContractMetadataState {
	baseStore := ctx.KVStore(s.key)
	return ContractMetadataState{
		stateStore: prefix.NewStore(baseStore, types.ContractMetadataStatePrefix),
		cdc:        s.cdc,
		ctx:        ctx,
	}
}
