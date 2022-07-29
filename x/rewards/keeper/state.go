package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/rewards/types"
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

// BlockRewardsState returns types.BlockRewards repository.
func (s State) BlockRewardsState(ctx sdk.Context) BlockRewardsState {
	baseStore := ctx.KVStore(s.key)
	return BlockRewardsState{
		stateStore: prefix.NewStore(baseStore, types.BlockRewardsStatePrefix),
		cdc:        s.cdc,
		ctx:        ctx,
	}
}

// TxRewardsState returns types.TxRewards repository.
func (s State) TxRewardsState(ctx sdk.Context) TxRewardsState {
	baseStore := ctx.KVStore(s.key)
	return TxRewardsState{
		stateStore: prefix.NewStore(baseStore, types.TxRewardsStatePrefix),
		cdc:        s.cdc,
		ctx:        ctx,
	}
}

// GetState returns the module storage state.
// Only for testing purposes.
func (k Keeper) GetState() State {
	return k.state
}
