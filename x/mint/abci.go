package mint

import (
	"time"

	"github.com/archway-network/archway/x/mint/keeper"
	"github.com/archway-network/archway/x/mint/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker mints new tokens and distributes to the inflation recipients.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	tokenToMint, blockInflation := k.GetBlockProvisions(ctx)

	// if no tokens to be minted
	if tokenToMint.IsZero() {
		return
	}

	// mint the tokens to the recipients
	mintAndDistribute(k, ctx, tokenToMint)

	// update the current block inflation
	blockTime := ctx.BlockTime()
	lbi := types.LastBlockInfo{
		Inflation: blockInflation,
		Time:      &blockTime,
	}
	err := k.SetLastBlockInfo(ctx, lbi)
	if err != nil {
		panic(err)
	}
}

func mintAndDistribute(k keeper.Keeper, ctx sdk.Context, tokenToMint sdk.Dec) {
	mintParams := k.GetParams(ctx)
	denom := k.BondDenom(ctx)
	mintCoin := sdk.NewInt64Coin(denom, tokenToMint.BigInt().Int64()) // as sdk.Coin

	err := k.MintCoins(ctx, sdk.NewCoins(mintCoin))
	if err != nil {
		panic(err)
	}

	for _, distribution := range mintParams.GetInflationRecipients() {
		amount := mintCoin.Amount.ToDec().Mul(distribution.Ratio) // totalCoinsToMint * distribution.Ratio
		coin := sdk.NewInt64Coin(denom, amount.BigInt().Int64())  // as sdk.Coin

		err := k.SendCoinsToModule(ctx, distribution.Recipient, sdk.NewCoins(coin))
		if err != nil {
			panic(err)
		}
		k.SetInflationForRecipient(ctx, distribution.Recipient, coin) // store how much was was minted for given module
	}
}
