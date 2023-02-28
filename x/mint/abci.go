package mint

import (
	"time"

	"github.com/archway-network/archway/x/mint/keeper"
	"github.com/archway-network/archway/x/mint/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const Year = 24 * time.Hour * 365

// BeginBlocker mints new tokens
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	mintParams := k.GetParams(ctx)
	lbi := getLastBlockInfo(ctx, k, mintParams)                             // get last block inflation and time
	elapsed := getTimeElapsed(ctx, lbi, mintParams)                         // time since last minting
	inflation := getCurrentBlockInflation(ctx, k, lbi, mintParams, elapsed) // inflation for the current block
	mintAmount := getAmountToMint(ctx, k, inflation, elapsed)               // amount of bond tokens to mint in this block
	mintInflation(ctx, k, mintAmount, mintParams)                           // minting and distributing inflation
	updateBlockInfo(ctx, k, inflation)                                      // updating blockinfo
}

// mintInflation mints the given amount of tokens and distributes to the
func mintInflation(ctx sdk.Context, k keeper.Keeper, totalCoinsToMint sdk.Dec, mintParams types.Params) {
	denom := k.BondDenom(ctx)
	for _, distribution := range mintParams.GetInflationRecipients() {
		amount := totalCoinsToMint.Mul(distribution.Ratio)       // totalCoinsToMint * distribution.Ratio
		coin := sdk.NewInt64Coin(denom, amount.BigInt().Int64()) // as sdk.Coin
		err := k.MintCoin(ctx, distribution.Recipient, coin)
		if err != nil {
			panic(err)
		}
	}
}

func getAmountToMint(ctx sdk.Context, k keeper.Keeper, inflation sdk.Dec, elapsed time.Duration) sdk.Dec {
	bondedTokenSupply := k.GetBondedTokenSupply(ctx)
	amount := inflation.MulInt(bondedTokenSupply.Amount).MulInt64(int64(elapsed / Year)) // amount := (inflation * bondedTokenSupply) * (elapsed/Year)
	return amount
}

func getCurrentBlockInflation(ctx sdk.Context, k keeper.Keeper, lbi types.LastBlockInfo, mintParams types.Params, elapsed time.Duration) sdk.Dec {
	bondedRatio := k.BondedRatio(ctx)
	inflation := getBlockInflation(lbi.Inflation, bondedRatio, mintParams, elapsed)
	return inflation
}

func getBlockInflation(inflation sdk.Dec, bondedRatio sdk.Dec, mintParams types.Params, elapsed time.Duration) sdk.Dec {
	switch {
	case bondedRatio.LT(mintParams.MinBonded): // if bondRatio is lower than we want, increase inflation
		inflation = inflation.Add(mintParams.InflationChange.MulInt64(int64(elapsed)))
	case bondedRatio.GT(mintParams.MaxBonded): // if bondRatio is higher than we want, decrease inflation
		inflation = inflation.Sub(mintParams.InflationChange.MulInt64(int64(elapsed)))
	}
	if inflation.GT(mintParams.MaxInflation) {
		inflation = mintParams.MaxInflation
	} else if inflation.LT(mintParams.MinInflation) {
		inflation = mintParams.MinInflation
	}
	return inflation
}

func getTimeElapsed(ctx sdk.Context, lbi types.LastBlockInfo, mintParams types.Params) time.Duration {
	elapsed := ctx.BlockTime().Sub(*lbi.GetTime()) // time since last mint
	if elapsed > mintParams.GetMaxBlockDuration() {
		elapsed = mintParams.GetMaxBlockDuration()
	}
	return elapsed
}

func getLastBlockInfo(ctx sdk.Context, k keeper.Keeper, mintParams types.Params) types.LastBlockInfo {
	lbi, found := k.GetLastBlockInfo(ctx)
	if !found {
		currentTime := ctx.BlockTime()
		lbi = types.LastBlockInfo{
			Inflation: mintParams.MinInflation,
			Time:      &currentTime,
		}
	}
	return lbi
}

func updateBlockInfo(ctx sdk.Context, k keeper.Keeper, inflation sdk.Dec) {
	blockTime := ctx.BlockTime()
	lbi := types.LastBlockInfo{
		Inflation: inflation,
		Time:      &blockTime,
	}
	err := k.SetLastBlockInfo(ctx, lbi)
	if err != nil {
		panic(err)
	}
}
