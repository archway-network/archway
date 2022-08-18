package pkg

import (
	"fmt"

	wasmVmTypes "github.com/CosmWasm/wasmvm/types"
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

// WasmCoinToSDK converts wasmVmTypes.Coin to sdk.Coin.
func WasmCoinToSDK(coin wasmVmTypes.Coin) (sdk.Coin, error) {
	amount, ok := sdk.NewIntFromString(coin.Amount)
	if !ok {
		return sdk.Coin{}, fmt.Errorf("invalid amount: %s", coin.Amount)
	}

	return sdk.Coin{
		Denom:  coin.Denom,
		Amount: amount,
	}, nil
}

// WasmCoinsToSDK converts wasmVmTypes.Coins to sdk.Coins.
func WasmCoinsToSDK(coins wasmVmTypes.Coins) (sdk.Coins, error) {
	result := make(sdk.Coins, 0, len(coins))
	for _, coin := range coins {
		coinSDK, err := WasmCoinToSDK(coin)
		if err != nil {
			return nil, err
		}
		result = append(result, coinSDK)
	}

	return result, nil
}
