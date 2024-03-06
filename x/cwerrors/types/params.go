package types

import (
	fmt "fmt"
)

var (
	DefaultErrorStoredTime       = int64(1000000)
	DefaultDisableErrorCallbacks = false
)

// NewParams creates a new Params instance.
func NewParams(
	errorStoredTime int64,
	disableErrorCallbacks bool,
) Params {
	return Params{
		ErrorStoredTime:       errorStoredTime,
		DisableErrorCallbacks: disableErrorCallbacks,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultErrorStoredTime,
		DefaultDisableErrorCallbacks,
	)
}

// Validate perform object fields validation.
func (p Params) Validate() error {
	if p.ErrorStoredTime <= 0 {
		return fmt.Errorf("ErrorStoredTime must be greater than 0. Current value: %d", p.ErrorStoredTime)
	}
	return nil
}
