package cwerrors

import (
	"github.com/archway-network/archway/x/cwerrors/keeper"
	"github.com/archway-network/archway/x/cwerrors/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the module genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	params := genState.Params
	if err := k.Params.Set(ctx, params); err != nil {
		panic(err)
	}
	if err := k.ErrorsCount.Set(ctx, 0); err != nil {
		panic(err)
	}
}

// ExportGenesis exports the module genesis for the current block.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(err)
	}
	return types.NewGenesisState(params)
}
