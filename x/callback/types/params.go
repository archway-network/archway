package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	DefaultCallbackGasLimit               = uint64(1000000)
	DefaultMaxBlockReservationLimit       = uint64(3)
	DefaultMaxFutureReservationLimit      = uint64(10000)
	DefaultBlockReservationFeeMultiplier  = sdk.MustNewDecFromStr("1.0")
	DefaultFutureReservationFeeMultiplier = sdk.MustNewDecFromStr("1.0")
)

// NewParams creates a new Params instance.
func NewParams(
	callbackGasLimit uint64,
	maxBlockReservationLimit uint64,
	maxFutureReservationLimit uint64,
	blockReservationFeeMultiplier sdk.Dec,
	futureReservationFeeMultiplier sdk.Dec,
) Params {
	return Params{
		CallbackGasLimit:               callbackGasLimit,
		MaxBlockReservationLimit:       maxBlockReservationLimit,
		MaxFutureReservationLimit:      maxFutureReservationLimit,
		BlockReservationFeeMultiplier:  blockReservationFeeMultiplier,
		FutureReservationFeeMultiplier: futureReservationFeeMultiplier,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultCallbackGasLimit,
		DefaultMaxBlockReservationLimit,
		DefaultMaxFutureReservationLimit,
		DefaultBlockReservationFeeMultiplier,
		DefaultFutureReservationFeeMultiplier,
	)
}

// Validate perform object fields validation.
func (p Params) Validate() error {
	if p.CallbackGasLimit == 0 {
		return fmt.Errorf("CallbackGasLimit must be greater than 0")
	}
	if p.BlockReservationFeeMultiplier.IsNegative() {
		return fmt.Errorf("BlockReservationFeeMultiplier must be greater than 0")
	}
	if p.FutureReservationFeeMultiplier.IsNegative() {
		return fmt.Errorf("FutureReservationFeeMultiplier must be greater than 0")
	}
	return nil
}
