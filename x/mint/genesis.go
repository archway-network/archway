package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/mint/keeper"
	"github.com/archway-network/archway/x/mint/types"
)

// InitGenesis initializes the module genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.GetParams())
	lbi := genState.GetLastBlockInfo()
	if (lbi == types.LastBlockInfo{}) {
		time := ctx.BlockTime()
		lbi = types.LastBlockInfo{
			Inflation: genState.Params.MinInflation,
			Time:      &time,
		}
	}
	if err := k.SetLastBlockInfo(ctx, lbi); err != nil {
		panic(err)
	}
}

// ExportGenesis exports the module genesis for the current block.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params := k.GetParams(ctx)
	lbi, found := k.GetLastBlockInfo(ctx)
	if !found {
		panic("could not find last block info")
	}
	return types.NewGenesisState(params, lbi)
}
