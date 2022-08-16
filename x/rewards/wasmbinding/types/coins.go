package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Coin is the WASM bindings version of the sdk.Coin.
type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// Coins is the WASM bindings version of the sdk.Coins.
type Coins []Coin

// NewCoinFromSDK converts the SDK version of the sdk.Coin to the WASM bindings version.
func NewCoinFromSDK(coin sdk.Coin) Coin {
	return Coin{
		Denom:  coin.Denom,
		Amount: coin.Amount.String(),
	}
}

// ToSDK converts the WASM bindings version of the sdk.Coin to the SDK version.
func (c Coin) ToSDK() (sdk.Coin, error) {
	amount, ok := sdk.NewIntFromString(c.Amount)
	if !ok {
		return sdk.Coin{}, fmt.Errorf("invalid amount: %s", c.Amount)
	}

	return sdk.Coin{
		Denom:  c.Denom,
		Amount: amount,
	}, nil
}

// NewCoinsFromSDK converts the SDK version of the sdk.Coins to the WASM bindings version.
func NewCoinsFromSDK(coins sdk.Coins) Coins {
	result := make(Coins, 0, len(coins))
	for _, coin := range coins {
		result = append(result, NewCoinFromSDK(coin))
	}

	return result
}

// ToSDK converts the WASM bindings version of the sdk.Coins to the SDK version.
func (c Coins) ToSDK() (sdk.Coins, error) {
	result := make(sdk.Coins, 0, len(c))
	for _, coin := range c {
		coinSDK, err := coin.ToSDK()
		if err != nil {
			return nil, err
		}
		result = append(result, coinSDK)
	}

	return result, nil
}
