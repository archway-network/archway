package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/pkg"
	"github.com/archway-network/archway/x/rewards/types"
)

// ExportGenesis exports the module genesis for the current block.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	minConsFee, _ := k.state.MinConsensusFee(ctx).GetFee() // default sdk.Coin value is ok
	rewardsRecordLastID, rewardsRecords := k.state.RewardsRecord(ctx).Export()

	var contractMetadata []types.ContractMetadata
	err := k.ContractMetadata.Walk(ctx, nil, func(key []byte, value types.ContractMetadata) (stop bool, err error) {
		contractMetadata = append(contractMetadata, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	return types.NewGenesisState(
		k.GetParams(ctx),
		contractMetadata,
		k.state.BlockRewardsState(ctx).Export(),
		k.state.TxRewardsState(ctx).Export(),
		minConsFee,
		rewardsRecordLastID,
		rewardsRecords,
		k.state.FlatFee(ctx).Export(),
	)
}

// InitGenesis initializes the module genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, state *types.GenesisState) {
	if err := k.SetParams(ctx, state.Params); err != nil {
		panic(err)
	}
	for _, contractMetadata := range state.ContractsMetadata {
		err := k.ContractMetadata.Set(ctx, contractMetadata.MustGetContractAddress(), contractMetadata)
		if err != nil {
			panic(err)
		}
	}
	k.state.BlockRewardsState(ctx).Import(state.BlockRewards)
	k.state.TxRewardsState(ctx).Import(state.TxRewards)
	k.state.RewardsRecord(ctx).Import(state.RewardsRecordLastId, state.RewardsRecords)
	k.state.FlatFee(ctx).Import(state.FlatFees)

	if !pkg.DecCoinIsZero(state.MinConsensusFee) && !pkg.DecCoinIsNegative(state.MinConsensusFee) {
		k.state.MinConsensusFee(ctx).SetFee(state.MinConsensusFee)
	}
}
