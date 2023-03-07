package mint

import (
	"github.com/archway-network/archway/x/mint/keeper"
	"github.com/archway-network/archway/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker mints new tokens and distributes to the inflation recipients.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	tokenToMint, blockInflation := k.GetBlockProvisions(ctx)

	// mint the tokens to the recipients
	err := k.MintCoins(ctx, "module", sdk.NewCoins(tokenToMint))
	if err != nil {
		panic(err)
	}

	// update the current block inflation
	blockTime := ctx.BlockTime()
	lbi := types.LastBlockInfo{
		Inflation: blockInflation,
		Time:      &blockTime,
	}
	err = k.SetLastBlockInfo(ctx, lbi)
	if err != nil {
		panic(err)
	}
}
