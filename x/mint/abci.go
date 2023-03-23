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

	// emit event of the mint amount and inflation
	types.EmitBlockInflationEvent(
		ctx,
		tokenToMint,
		blockInflation,
	)

	// mint the tokens to the recipients
	err := mintAndDistribute(k, ctx, tokenToMint)
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

func mintAndDistribute(k keeper.Keeper, ctx sdk.Context, tokensToMint sdk.Dec) error {
	mintParams := k.GetParams(ctx)
	denom := k.BondDenom(ctx)

	for _, distribution := range mintParams.GetInflationRecipients() {
		amount := distribution.Ratio.Mul(tokensToMint)   // distribution.Ratio * totalMintCoins
		coin := sdk.NewCoin(denom, amount.TruncateInt()) // as sdk.Coin

		err := k.MintCoins(ctx, sdk.NewCoins(coin)) // mint the tokens into x/mint
		if err != nil {
			return err
		}

		err = k.SendCoinsToModule(ctx, distribution.Recipient, sdk.NewCoins(coin)) // distribute the tokens from x/mint
		if err != nil {
			return err
		}
		k.SetInflationForRecipient(ctx, distribution.Recipient, coin) // store how much was was minted for given module

		types.EmitBlockInflationDistributionEvent( // emit event of the mint distribution
			ctx,
			distribution.Recipient,
			coin,
		)
	}
	return nil
}
