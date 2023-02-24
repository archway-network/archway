package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewParams creates a new Params instance.
func NewParams(inflationRewardsRatio, txFeeRebateRatio sdk.Dec, maxwithdrawRecords uint64) Params {
	panic("unimplemented 👻")
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	panic("unimplemented 👻")
}

// Validate perform object fields validation.
func (m Params) Validate() error {
	if err := validateInflation(m.MinInflation); err != nil {
		return sdkErrors.Wrap(err, "min_inflation param has invalid value, should be between 0 and 1")
	}
	if err := validateInflation(m.MaxInflation); err != nil {
		return sdkErrors.Wrap(err, "max_inflation param has invalid value, should be between 0 and 1")
	}
	if err := validateBondedRatio(m.MinBonded); err != nil {
		return sdkErrors.Wrap(err, "min_bonded param has invalid value, should be between 0 and 1")
	}
	if err := validateBondedRatio(m.MaxBonded); err != nil {
		return sdkErrors.Wrap(err, "max_bonded param has invalid value, should be between 0 and 1")
	}
	if err := validateInflationChange(m.InflationChange); err != nil {
		return sdkErrors.Wrap(err, "inflation_change param has invalid value, should be between 0 and 1")
	}
	if err := validateInflationRecipients(m.InflationRecipients); err != nil {
		return err
	}
	return nil
}

func validateInflation(inflation sdk.Dec) error {
	if !isDecValidPercentage(inflation) {
		return ErrInvalidInflation
	}
	return nil
}

func validateBondedRatio(bondedRatio sdk.Dec) error {
	if !isDecValidPercentage(bondedRatio) {
		return ErrInvalidBondedRatio
	}
	return nil
}

func validateInflationChange(inflationChange sdk.Dec) error {
	if !isDecValidPercentage(inflationChange) {
		return ErrInvalidInflationChange
	}
	return nil
}

func validateInflationRecipients(inflationRecipients []*Params_InflationRecipient) error {
	if len(inflationRecipients) < 1 {
		return sdkErrors.Wrap(ErrInvalidInflationRecipient, "inflation recipients not found")
	}
	inflationDistribution := sdk.ZeroDec()
	for _, recipient := range inflationRecipients {
		inflationDistribution = inflationDistribution.Add(recipient.Ratio)
	}
	if !inflationDistribution.Equal(sdk.OneDec()) {
		return sdkErrors.Wrap(ErrInvalidInflationDistribution, "inflation distribution sum is not equal to 1")
	}
	return nil
}

// isDecValidPercentage returns true if the given sdk.Dec value is between 0 and 1
func isDecValidPercentage(dec sdk.Dec) bool {
	if dec.LT(sdk.ZeroDec()) || dec.GT(sdk.OneDec()) {
		return false
	}
	return true
}
