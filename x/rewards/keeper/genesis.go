package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/pkg"
	"github.com/archway-network/archway/x/rewards/types"
)

// ExportGenesis exports the module genesis for the current block.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	minConsFee, _ := k.MinConsFee.Get(ctx) // default sdk.Coin value is ok

	var contractMetadata []types.ContractMetadata
	err := k.ContractMetadata.Walk(ctx, nil, func(key []byte, value types.ContractMetadata) (stop bool, err error) {
		contractMetadata = append(contractMetadata, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	var flatFees []types.FlatFee
	err = k.FlatFees.Walk(ctx, nil, func(key []byte, value sdk.Coin) (stop bool, err error) {
		flatFees = append(flatFees, types.FlatFee{
			ContractAddress: sdk.AccAddress(key).String(),
			FlatFee:         value,
		})
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	var blockRewards []types.BlockRewards
	err = k.BlockRewards.Walk(ctx, nil, func(key uint64, value types.BlockRewards) (stop bool, err error) {
		blockRewards = append(blockRewards, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	var txRewards []types.TxRewards
	err = k.TxRewards.Walk(ctx, nil, func(_ uint64, value types.TxRewards) (stop bool, err error) {
		txRewards = append(txRewards, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	var rewardsRecords []types.RewardsRecord
	err = k.RewardsRecords.Walk(ctx, nil, func(key uint64, value types.RewardsRecord) (stop bool, err error) {
		rewardsRecords = append(rewardsRecords, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	rewardsRecordLastID, err := k.RewardsRecordID.Peek(ctx)
	if err != nil {
		panic(err)
	}

	return types.NewGenesisState(
		k.GetParams(ctx),
		contractMetadata,
		blockRewards,
		txRewards,
		minConsFee,
		rewardsRecordLastID,
		rewardsRecords,
		flatFees,
	)
}

// InitGenesis initializes the module genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, state *types.GenesisState) {
	if err := k.Params.Set(ctx, state.Params); err != nil {
		panic(err)
	}
	for _, contractMetadata := range state.ContractsMetadata {
		err := k.ContractMetadata.Set(ctx, contractMetadata.MustGetContractAddress(), contractMetadata)
		if err != nil {
			panic(err)
		}
	}

	for _, flatFee := range state.FlatFees {
		err := k.FlatFees.Set(ctx, flatFee.MustGetContractAddress(), flatFee.FlatFee)
		if err != nil {
			panic(err)
		}
	}

	for _, blockReward := range state.BlockRewards {
		err := k.BlockRewards.Set(ctx, uint64(blockReward.Height), blockReward)
		if err != nil {
			panic(err)
		}
	}

	for _, txReward := range state.TxRewards {
		err := k.TxRewards.Set(ctx, uint64(txReward.Height), txReward)
		if err != nil {
			panic(err)
		}
	}

	for _, rewardsRecord := range state.RewardsRecords {
		err := k.RewardsRecords.Set(ctx, rewardsRecord.Id, rewardsRecord)
		if err != nil {
			panic(err)
		}
	}
	err := k.RewardsRecordID.Set(ctx, state.RewardsRecordLastId)
	if err != nil {
		panic(err)
	}

	if !pkg.DecCoinIsZero(state.MinConsensusFee) && !pkg.DecCoinIsNegative(state.MinConsensusFee) {
		err := k.MinConsFee.Set(ctx, state.MinConsensusFee)
		if err != nil {
			panic(err)
		}
	}
}
