package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	DefaultErrorStoredTime    = int64(302400)                             // roughly 21 days
	DefaultSubscriptionFee    = sdk.NewInt64Coin(sdk.DefaultBondDenom, 0) // 1 ARCH (1e18 attoarch)
	DefaultSubscriptionPeriod = int64(302400)                             // roughly 21 days
)

// NewParams creates a new Params instance.
func NewParams(
	errorStoredTime int64,
	subscriptionFee sdk.Coin,
	subscriptionPeriod int64,
) Params {
	return Params{
		ErrorStoredTime:    errorStoredTime,
		SubscriptionFee:    subscriptionFee,
		SubscriptionPeriod: subscriptionPeriod,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultErrorStoredTime,
		DefaultSubscriptionFee,
		DefaultSubscriptionPeriod,
	)
}

// Validate perform object fields validation.
func (p Params) Validate() error {
	if p.ErrorStoredTime <= 0 {
		return fmt.Errorf("ErrorStoredTime must be greater than 0. Current value: %d", p.ErrorStoredTime)
	}
	if !p.SubscriptionFee.IsValid() {
		return fmt.Errorf("SubsciptionFee is not valid. Current value: %s", p.SubscriptionFee)
	}
	if p.SubscriptionPeriod <= 0 {
		return fmt.Errorf("SubscriptionPeriod must be greater than 0. Current value: %d", p.SubscriptionPeriod)
	}
	return nil
}
