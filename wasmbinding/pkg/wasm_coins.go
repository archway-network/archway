package pkg

import (
	"fmt"

	wasmVmTypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

// SDKCoinToWasm converts sdk.Coin to wasmVmTypes.Coin
func SDKCoinToWasm(coin sdk.Coin) wasmVmTypes.Coin {
	wasmCoin := wasmVmTypes.Coin{
		Denom:  coin.Denom,
		Amount: coin.Amount.String(),
	}
	return wasmCoin
}
