package cwerrors

import (
	"github.com/archway-network/archway/x/cwerrors/keeper"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker is called every block, and prunes errors that are older than the current block height.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	err := k.PruneErrorsByBlockHeight(ctx, ctx.BlockHeight())
	if err != nil {
		panic(err)
	}
	return nil
}
