package cwerrors

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/cwerrors/keeper"
	"github.com/archway-network/archway/x/cwerrors/types"
)

// InitGenesis initializes the module genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	params := genState.Params
	if err := k.Params.Set(ctx, params); err != nil {
		panic(err)
	}
}

// ExportGenesis exports the module genesis for the current block.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(err)
	}
	sudoErrs, err := k.ExportErrors(ctx)
	if err != nil {
		panic(err)
	}
	genesis := types.NewGenesisState(params)
	genesis.Errors = sudoErrs
	return genesis
}
