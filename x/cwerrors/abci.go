package callback

import (
	"github.com/archway-network/archway/x/cwerrors/keeper"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker fetches all the callbacks registered for the current block height and executes them
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	err := k.PruneErrorsByBlockHeight(ctx, ctx.BlockHeight())
	if err != nil {
		panic(err)
	}
	return nil
}
