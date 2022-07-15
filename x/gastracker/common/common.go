package common

import sdk "github.com/cosmos/cosmos-sdk/types"

func SplitCoins(ratio sdk.Dec, fees sdk.Coins) (authCoins, gasTrackerCoins sdk.Coins) {
	authCoins = make(sdk.Coins, len(fees))
	gasTrackerCoins = make(sdk.Coins, len(fees))

	for i, feeCoin := range fees {
		gasTrackerCoin := sdk.Coin{
			Denom:  feeCoin.Denom,
			Amount: feeCoin.Amount.ToDec().Mul(ratio).TruncateInt(),
		}

		gasTrackerCoins[i] = gasTrackerCoin
		authCoins[i] = feeCoin.Sub(gasTrackerCoin)
	}

	return authCoins, gasTrackerCoins
}
