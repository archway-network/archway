package custodian

import (
	"github.com/archway-network/archway/x/custodian/keeper"
	"github.com/archway-network/archway/x/custodian/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the custodian module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	err := k.SetParams(ctx, genState.Params)
	if err != nil {
		panic(err)
	}
}

// ExportGenesis returns the custodian module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	return genesis
}
