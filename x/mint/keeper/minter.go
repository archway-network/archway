package keeper

import (
	"math/big"
	"time"

	"github.com/archway-network/archway/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const Year = 24 * time.Hour * 365

// GetInflationForRecipient gets the sdk.Coin distributed to the given module in the current block
func (k Keeper) GetInflationForRecipient(ctx sdk.Context, recipientName string) (sdk.Coin, bool) {
	store := ctx.KVStore(k.storeKey)

	var mintAmount sdk.Coin
	bz := store.Get(types.GetMintDistributionRecipientKey(ctx.BlockHeight(), recipientName))
	if bz == nil {
		return mintAmount, false
	}

	k.cdc.MustUnmarshal(bz, &mintAmount)
	return mintAmount, true
}

// SetInflationForRecipient sets the sdk.Coin distributed to the given module for the current block
func (k Keeper) SetInflationForRecipient(ctx sdk.Context, recipientName string, mintAmount sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	value := k.cdc.MustMarshal(&mintAmount)

	store.Set(types.GetMintDistributionRecipientKey(ctx.BlockHeight(), recipientName), value)
}

// GetBlockProvisions gets the tokens to be minted in the current block and returns the new inflation amount as well
func (k Keeper) GetBlockProvisions(ctx sdk.Context) (tokens sdk.Dec, blockInflation sdk.Dec) {
	mintParams := k.GetParams(ctx)

	// getting last block info
	lbi, found := k.GetLastBlockInfo(ctx)
	if !found {
		currentTime := ctx.BlockTime()
		lbi = types.LastBlockInfo{
			Inflation: mintParams.MinInflation,
			Time:      &currentTime,
		}
	}

	// time since last mint
	elapsed := ctx.BlockTime().Sub(*lbi.GetTime())
	if maxElapsed := mintParams.GetMaxBlockDuration(); elapsed > maxElapsed {
		elapsed = maxElapsed
	}

	// inflation for the current block
	bondedRatio := k.BondedRatio(ctx)
	blockInflation = getBlockInflation(lbi.Inflation, bondedRatio, mintParams, elapsed)

	// amount of bond tokens to mint in this block
	bondedTokenSupply := k.GetBondedTokenSupply(ctx)

	tokens = blockInflation.MulInt(bondedTokenSupply.Amount).Mul(sdk.NewDecFromBigInt(big.NewInt(int64(elapsed.Seconds()))).QuoInt64(int64(Year.Seconds()))) // amount := (inflation * bondedTokenSupply) * (elapsed/Year)
	return
}

// getBlockInflation adjusts the current block inflation amount based on tokens bonded
func getBlockInflation(inflation sdk.Dec, bondedRatio sdk.Dec, mintParams types.Params, elapsed time.Duration) sdk.Dec {
	switch {
	case bondedRatio.LT(mintParams.MinBonded): // if bondRatio is lower than we want, increase inflation
		inflation = inflation.Add(calculateInflationChange(mintParams, elapsed))
	case bondedRatio.GT(mintParams.MaxBonded): // if bondRatio is higher than we want, decrease inflation
		inflation = inflation.Sub(calculateInflationChange(mintParams, elapsed))
	}
	if inflation.GT(mintParams.MaxInflation) {
		inflation = mintParams.MaxInflation
	} else if inflation.LT(mintParams.MinInflation) {
		inflation = mintParams.MinInflation
	}
	return inflation
}

func calculateInflationChange(mintParams types.Params, elapsed time.Duration) sdk.Dec {
	return mintParams.InflationChange.MulInt64(int64(elapsed.Seconds()))
}
