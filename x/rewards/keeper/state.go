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

// DeleteBlockRewardsCascade deletes all block rewards for a given height.
// Function removes BlockRewards and TxRewards objects cleaning up their indexes.
func (s State) DeleteBlockRewardsCascade(ctx sdk.Context, height int64) {
	s.BlockRewardsState(ctx).DeleteBlockRewards(height)
	s.TxRewardsState(ctx).deleteTxRewardsByBlock(height)
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

// MinConsensusFee returns the Minimum Consensus Fee repository.
func (s State) MinConsensusFee(ctx sdk.Context) MinConsFeeState {
	baseStore := ctx.KVStore(s.key)
	return MinConsFeeState{
		stateStore: prefix.NewStore(baseStore, types.MinConsFeeStatePrefix),
		cdc:        s.cdc,
		ctx:        ctx,
	}
}

// RewardsRecord returns types.RewardsRecord repository.
func (s State) RewardsRecord(ctx sdk.Context) RewardsRecordState {
	baseStore := ctx.KVStore(s.key)
	return RewardsRecordState{
		stateStore: prefix.NewStore(baseStore, types.RewardsRecordStatePrefix),
		cdc:        s.cdc,
		ctx:        ctx,
	}
}

// FlatFee returns types.FlatFeeState repository.
func (s State) FlatFee(ctx sdk.Context) FlatFeeState {
	baseStore := ctx.KVStore(s.key)
	return FlatFeeState{
		stateStore: prefix.NewStore(baseStore, types.FlatFeeStatePrefix),
		cdc:        s.cdc,
		ctx:        ctx,
	}
}

// GetState returns the module storage state.
// Only for testing purposes.
func (k Keeper) GetState() State {
	return k.state
}
