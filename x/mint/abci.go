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
	mintInflation(ctx, k, mintAmount)                                       // minting and distributing inflation
	updateBlockInfo(ctx, k, inflation)                                      // updating blockinfo
}

// mintInflation mints the given amount of tokens and distributes to the
func mintInflation(ctx sdk.Context, k keeper.Keeper, mintAmount sdk.Dec) {
	err := k.MintCoin(ctx, types.ModuleName, sdk.NewInt64Coin(k.BondDenom(ctx), mintAmount.BigInt().Int64()))
	if err != nil {
		panic(err)
	}
}

func getCurrentBlockInflation(ctx sdk.Context, k keeper.Keeper, lbi types.LastBlockInfo, mintParams types.Params, elapsed time.Duration) sdk.Dec {
	bondedRatio := k.BondedRatio(ctx)
	inflation := getBlockInflation(lbi.Inflation, bondedRatio, mintParams, elapsed)
	return inflation
}

func getAmountToMint(ctx sdk.Context, k keeper.Keeper, inflation sdk.Dec, elapsed time.Duration) sdk.Dec {
	bondedTokenSupply := k.GetBondedTokenSupply(ctx)
	amount := inflation.MulInt(bondedTokenSupply.Amount).MulInt64(int64(elapsed / Year))
	return amount
}

func getBlockInflation(inflation sdk.Dec, bondedRatio sdk.Dec, mintParams types.Params, elapsed time.Duration) sdk.Dec {
	switch {
	case bondedRatio.LT(mintParams.MinBonded):
		inflation = inflation.Add(mintParams.InflationChange.MulInt64(int64(elapsed)))
	case bondedRatio.GT(mintParams.MaxBonded):
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
	elapsed := ctx.BlockTime().Sub(*lbi.GetTime())
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
