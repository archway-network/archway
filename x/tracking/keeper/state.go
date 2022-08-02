package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/tracking/types"
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

// DeleteTxInfosCascade deletes all block tracking for a given height.
// Function removes TxInfo and ContractOpInfo objects cleaning up their indexes.
func (s State) DeleteTxInfosCascade(ctx sdk.Context, height int64) {
	contractOpInfoState := s.ContractOpInfoState(ctx)

	txIDs := s.TxInfoState(ctx).DeleteTxInfosByBlock(height)
	for _, txID := range txIDs {
		contractOpInfoState.DeleteContractOpsByTxID(txID)
	}
}

// TxInfoState returns types.TxInfo repository.
func (s State) TxInfoState(ctx sdk.Context) TxInfoState {
	baseStore := ctx.KVStore(s.key)
	return TxInfoState{
		stateStore: prefix.NewStore(baseStore, types.TxInfoStatePrefix),
		cdc:        s.cdc,
		ctx:        ctx,
	}
}

// ContractOpInfoState returns types.ContractOperationInfo repository.
func (s State) ContractOpInfoState(ctx sdk.Context) ContractOpInfoState {
	baseStore := ctx.KVStore(s.key)
	return ContractOpInfoState{
		stateStore: prefix.NewStore(baseStore, types.ContractOpInfoStatePrefix),
		cdc:        s.cdc,
		ctx:        ctx,
	}
}

// GetState returns the module storage state.
// Only for testing purposes.
func (k Keeper) GetState() State {
	return k.state
}
