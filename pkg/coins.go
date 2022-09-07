package pkg

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SplitCoins splits coins in a proportion defined by the ratio.
// CONTRACT: inputs must be valid.
func SplitCoins(coins sdk.Coins, ratio sdk.Dec) (stack1, stack2 sdk.Coins) {
	stack1 = sdk.NewCoins()
	stack2 = sdk.NewCoins()

	for _, coin := range coins {
		stack1Coin := sdk.Coin{
			Denom:  coin.Denom,
			Amount: coin.Amount.ToDec().Mul(ratio).TruncateInt(),
		}
		stack2Coin := coin.Sub(stack1Coin)

		stack1 = stack1.Add(stack1Coin)
		stack2 = stack2.Add(stack2Coin)
	}

	return
}
