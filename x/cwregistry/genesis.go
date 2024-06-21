package cwica

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/cwregistry/keeper"
	"github.com/archway-network/archway/x/cwregistry/types"
)

// InitGenesis initializes the cwregistry module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	panic("unimplemented 👻")
}

// ExportGenesis returns the cwregistry module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	panic("unimplemented 👻")
}
