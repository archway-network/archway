package types

import (
	"fmt"

	math "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramTypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	InflationRewardsRatioParamKey = []byte("InflationRewardsRatio")
	TxFeeRebateRatioParamKey      = []byte("TxFeeRebateRatio")
	MaxWithdrawRecordsParamKey    = []byte("MaxWithdrawRecords")
	MinPriceOfGasParamKey         = []byte("MinPriceOfGas")
)

// Limit below are var (not const) for E2E tests to change them.
var (
	// MaxWithdrawRecordsParamLimit defines the MaxWithdrawRecordsParamKey max value.
	// Limit is estimated by the TestRewardsParamMaxWithdrawRecordsLimit E2E test.
	MaxWithdrawRecordsParamLimit = uint64(25000) // limit is defined by the TestRewardsParamMaxWithdrawRecordsLimit E2E test
	// MaxRecordsQueryLimit defines the page limit for querying RewardsRecords.
	// Limit is defined by the TestRewardsRecordsQueryLimit E2E test.
	MaxRecordsQueryLimit = uint64(7500)
)

var (
	DefaultInflationRatio     = math.LegacyMustNewDecFromStr("0.20") // 20%
	DefaultTxFeeRebateRatio   = math.LegacyMustNewDecFromStr("0.50") // 50%
	DefaultMaxWithdrawRecords = MaxWithdrawRecordsParamLimit
	DefaultMinPriceOfGas      = sdk.NewDecCoin("stake", math.ZeroInt())
)

var _ paramTypes.ParamSet = (*Params)(nil)

// ParamKeyTable creates a new params table.
func ParamKeyTable() paramTypes.KeyTable {
	return paramTypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance.
func NewParams(inflationRewardsRatio, txFeeRebateRatio math.LegacyDec, maxwithdrawRecords uint64, minPriceOfGas sdk.DecCoin) Params {
	return Params{
		InflationRewardsRatio: inflationRewardsRatio,
		TxFeeRebateRatio:      txFeeRebateRatio,
		MaxWithdrawRecords:    maxwithdrawRecords,
		MinPriceOfGas:         minPriceOfGas,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultInflationRatio,
		DefaultTxFeeRebateRatio,
		DefaultMaxWithdrawRecords,
		DefaultMinPriceOfGas,
	)
}

// ParamSetPairs Implements the paramTypes.ParamSet interface.
func (m *Params) ParamSetPairs() paramTypes.ParamSetPairs {
	return paramTypes.ParamSetPairs{
		paramTypes.NewParamSetPair(InflationRewardsRatioParamKey, &m.InflationRewardsRatio, validateInflationRewardsRatio),
		paramTypes.NewParamSetPair(TxFeeRebateRatioParamKey, &m.TxFeeRebateRatio, validateTxFeeRebateRatio),
		paramTypes.NewParamSetPair(MaxWithdrawRecordsParamKey, &m.MaxWithdrawRecords, validateMaxWithdrawRecords),
		paramTypes.NewParamSetPair(MinPriceOfGasParamKey, &m.MinPriceOfGas, validateMinPriceOfGas),
	}
}

// Validate perform object fields validation.
func (m Params) Validate() error {
	if err := validateInflationRewardsRatio(m.InflationRewardsRatio); err != nil {
		return err
	}
	if err := validateTxFeeRebateRatio(m.TxFeeRebateRatio); err != nil {
		return err
	}
	if err := validateMaxWithdrawRecords(m.MaxWithdrawRecords); err != nil {
		return err
	}
	if err := validateMinPriceOfGas(m.MinPriceOfGas); err != nil {
		return err
	}
	return nil
}

func validateInflationRewardsRatio(v interface{}) (retErr error) {
	defer func() {
		if retErr != nil {
			retErr = fmt.Errorf("inflationRewardsRatio param: %w", retErr)
		}
	}()

	p, ok := v.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	return validateRatio(p)
}

func validateTxFeeRebateRatio(v interface{}) (retErr error) {
	defer func() {
		if retErr != nil {
			retErr = fmt.Errorf("txFeeRebateRatio param: %w", retErr)
		}
	}()

	p, ok := v.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	return validateRatio(p)
}

// validateRatio is a generic ratio coefficient validator.
func validateRatio(v math.LegacyDec) error {
	if v.IsNegative() {
		return fmt.Errorf("must be GTE 0.0")
	}
	if v.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("must be LT 1.0")
	}

	return nil
}

func validateMaxWithdrawRecords(v interface{}) (retErr error) {
	defer func() {
		if retErr != nil {
			retErr = fmt.Errorf("maxWithdrawRecords param: %w", retErr)
		}
	}()

	p, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if p == 0 {
		return fmt.Errorf("must be GTE 1")
	}
	if p > MaxWithdrawRecordsParamLimit {
		return fmt.Errorf("must be LTE %d", MaxWithdrawRecordsParamLimit)
	}

	return nil
}

func validateMinPriceOfGas(v interface{}) (retErr error) {
	defer func() {
		if retErr != nil {
			retErr = fmt.Errorf("minPriceOfGas param: %w", retErr)
		}
	}()

	p, ok := v.(sdk.DecCoin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	return p.Validate()
}
