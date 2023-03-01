package types

import (
	fmt "fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

// Default param values
var (
	DefaultMinimumInflation      sdk.Dec            = sdk.ZeroDec()
	DefaultMaximumInflation      sdk.Dec            = sdk.OneDec()
	DefaultMinimumBonded         sdk.Dec            = sdk.ZeroDec()
	DefaultMaximumBonded         sdk.Dec            = sdk.OneDec()
	DefaultInflationChange       sdk.Dec            = sdk.OneDec()
	DefaultMaxBlockDuration      time.Duration      = time.Minute
	DefaultFeeCollectorRecipient InflationRecipient = InflationRecipient{
		Recipient: authtypes.FeeCollectorName,
		Ratio:     sdk.OneDec(),
	}
)

// Parameter store keys.
var (
	KeyMinimumInflation    = []byte("MinimumInflation")
	KeyMaximumInflation    = []byte("MaximumInflation")
	KeyMinimumBonded       = []byte("MinimumBonded")
	KeyMaximumBonded       = []byte("MaximumBonded")
	KeyInflationChange     = []byte("InflationChange")
	KeyMaxBlockDuration    = []byte("MaxBlockDuration")
	KeyInflationRecipients = []byte("InflationRecipients")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// NewParams creates a new Params instance.
func NewParams(minInflation sdk.Dec, maxInflation sdk.Dec, minBonded sdk.Dec, maxBonded sdk.Dec, inflationChange sdk.Dec, maxBlockDuration time.Duration, inflationRecipients []*InflationRecipient) Params {
	return Params{
		MinInflation:        minInflation,
		MaxInflation:        maxInflation,
		MinBonded:           minBonded,
		MaxBonded:           maxBonded,
		InflationChange:     inflationChange,
		MaxBlockDuration:    maxBlockDuration,
		InflationRecipients: inflationRecipients,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultMinimumInflation,
		DefaultMaximumInflation,
		DefaultMinimumBonded,
		DefaultMaximumBonded,
		DefaultInflationChange,
		DefaultMaxBlockDuration,
		[]*InflationRecipient{&DefaultFeeCollectorRecipient},
	)
}

// ParamKeyTable the param key table for mint module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMinimumInflation, &p.MinInflation, validateInflation),
		paramtypes.NewParamSetPair(KeyMaximumInflation, &p.MaxInflation, validateInflation),
		paramtypes.NewParamSetPair(KeyMinimumBonded, &p.MinBonded, validateBondedRatio),
		paramtypes.NewParamSetPair(KeyMaximumBonded, &p.MaxBonded, validateBondedRatio),
		paramtypes.NewParamSetPair(KeyInflationChange, &p.InflationChange, validateInflationChange),
		paramtypes.NewParamSetPair(KeyMaxBlockDuration, &p.MaxBlockDuration, validateMaxBlockDuration),
		paramtypes.NewParamSetPair(KeyInflationRecipients, &p.InflationRecipients, validateInflationRecipients),
	}
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate perform object fields validation.
func (p Params) Validate() error {
	if err := validateInflation(p.MinInflation); err != nil {
		return sdkErrors.Wrap(err, "min_inflation param has invalid value, should be between 0 and 1")
	}
	if err := validateInflation(p.MaxInflation); err != nil {
		return sdkErrors.Wrap(err, "max_inflation param has invalid value, should be between 0 and 1")
	}
	if err := validateBondedRatio(p.MinBonded); err != nil {
		return sdkErrors.Wrap(err, "min_bonded param has invalid value, should be between 0 and 1")
	}
	if err := validateBondedRatio(p.MaxBonded); err != nil {
		return sdkErrors.Wrap(err, "max_bonded param has invalid value, should be between 0 and 1")
	}
	if err := validateInflationChange(p.InflationChange); err != nil {
		return sdkErrors.Wrap(err, "inflation_change param has invalid value, should be between 0 and 1")
	}
	if err := validateMaxBlockDuration(p.MaxBlockDuration); err != nil {
		return sdkErrors.Wrap(err, "max_block_duration param has invalid value, should be greater than 0")
	}
	if err := validateInflationRecipients(p.InflationRecipients); err != nil {
		return err
	}
	return nil
}

func validateInflation(i interface{}) error {
	inflation, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if !isDecValidPercentage(inflation) {
		return ErrInvalidInflation
	}
	return nil
}

func validateBondedRatio(i interface{}) error {
	bondedRatio, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if !isDecValidPercentage(bondedRatio) {
		return ErrInvalidBondedRatio
	}
	return nil
}

func validateInflationChange(i interface{}) error {
	inflationChange, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if !isDecValidPercentage(inflationChange) {
		return ErrInvalidInflationChange
	}
	return nil
}

func validateMaxBlockDuration(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("max block duration must be positive")
	}

	return nil
}

func validateInflationRecipients(i interface{}) error {
	inflationRecipients, ok := i.([]*InflationRecipient)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if len(inflationRecipients) < 1 {
		return sdkErrors.Wrap(ErrInvalidInflationRecipient, "inflation recipients not found")
	}
	inflationDistribution := sdk.ZeroDec()
	for _, recipient := range inflationRecipients {
		if recipient.Recipient == "" {
			return sdkErrors.Wrap(ErrInvalidInflationRecipient, "inflation recipient module name is empty")
		}
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
