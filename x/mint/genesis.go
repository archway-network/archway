package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/mint/keeper"
	"github.com/archway-network/archway/x/mint/types"
)

// InitGenesis initializes the module genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	panic("unimplemented👻")
}

// ExportGenesis exports the module genesis for the current block.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	panic("unimplemented👻")
}
