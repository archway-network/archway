package pkg

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CoinIsZero checks if sdk.Coin is set (not panics in case Amount is nil).
func CoinIsZero(coin sdk.Coin) bool {
	if coin.Amount.IsNil() {
		return true
	}

	return coin.IsZero()
}

// DecCoinIsZero checks if sdk.DecCoin is set (not panics in case Amount is nil).
func DecCoinIsZero(coin sdk.DecCoin) bool {
	if coin.Amount.IsNil() {
		return true
	}

	return coin.IsZero()
}

// DecCoinIsNegative checks if sdk.DecCoin is negative (not panics in case Amount is nil).
func DecCoinIsNegative(coin sdk.DecCoin) bool {
	if coin.Amount.IsNil() {
		return true
	}

	return coin.IsNegative()
}

// ValidateCoin performs a stricter validation of sdk.Coin comparing to the SDK version.
func ValidateCoin(coin sdk.Coin) error {
	if err := sdk.ValidateDenom(coin.Denom); err != nil {
		return fmt.Errorf("denom: %w", err)
	}
	if coin.Amount.IsNil() {
		return fmt.Errorf("amount: nil")
	}
	if coin.IsNegative() {
		return fmt.Errorf("amount: is negative")
	}

	return nil
}

// ValidateDecCoin performs a stricter validation of sdk.DecCoin comparing to the SDK version.
func ValidateDecCoin(coin sdk.DecCoin) error {
	if err := sdk.ValidateDenom(coin.Denom); err != nil {
		return fmt.Errorf("denom: %w", err)
	}
	if coin.Amount.IsNil() {
		return fmt.Errorf("amount: nil")
	}
	if coin.IsNegative() {
		return fmt.Errorf("amount: is negative")
	}

	return nil
}
