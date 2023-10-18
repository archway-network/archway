package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/rewards/types"
)

// State is a wrapper around the module storage state.
type State struct {
	key storetypes.StoreKey
	cdc codec.Codec
}

// NewState creates a new State instance.
func NewState(cdc codec.Codec, key storetypes.StoreKey) State {
	return State{
		key: key,
		cdc: cdc,
	}
}

// RewardsRecord returns types.RewardsRecord repository.
func (s State) RewardsRecord(ctx sdk.Context) RewardsRecordState {
	baseStore := ctx.KVStore(s.key)
	return RewardsRecordState{
		stateStore: prefix.NewStore(baseStore, types.RewardsRecordStatePrefix),
		cdc:        s.cdc,
	}
}

// GetState returns the module storage state.
// Only for testing purposes.
func (k Keeper) GetState() State {
	return k.state
}
