package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/mint/keeper"
	"github.com/archway-network/archway/x/mint/types"
)

// InitGenesis initializes the module genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.GetParams())
	k.SetLastBlockInfo(ctx, genState.GetLastBlockInfo())
}

// ExportGenesis exports the module genesis for the current block.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params := k.GetParams(ctx)
	lbi, found := k.GetLastBlockInfo(ctx)
	if !found {
		currentTime := ctx.BlockTime()
		lbi = types.LastBlockInfo{
			Inflation: params.MinInflation,
			Time:      &currentTime,
		}
	}
	return types.NewGenesisState(params, lbi)
}
