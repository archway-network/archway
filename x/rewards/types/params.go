package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramTypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"sigs.k8s.io/yaml"
)

var (
	InflationRewardsRatioParamKey = []byte("InflationRewardsRatio")
	TxFeeRebateRatioParamKey      = []byte("TxFeeRebateRatio")
)

var _ paramTypes.ParamSet = (*Params)(nil)

// ParamKeyTable creates a new params table.
func ParamKeyTable() paramTypes.KeyTable {
	return paramTypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance.
func NewParams(inflationRewardsRatio, txFeeRebateRatio sdk.Dec) Params {
	return Params{
		InflationRewardsRatio: inflationRewardsRatio,
		TxFeeRebateRatio:      txFeeRebateRatio,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	defInflationRatio := sdk.MustNewDecFromStr("0.20")   // 20%
	defTxFeeRebateRatio := sdk.MustNewDecFromStr("0.50") // 50%

	return NewParams(defInflationRatio, defTxFeeRebateRatio)
}

// ParamSetPairs Implements the paramTypes.ParamSet interface.
func (m *Params) ParamSetPairs() paramTypes.ParamSetPairs {
	return paramTypes.ParamSetPairs{
		paramTypes.NewParamSetPair(InflationRewardsRatioParamKey, &m.InflationRewardsRatio, validateInflationRewardsRatio),
		paramTypes.NewParamSetPair(TxFeeRebateRatioParamKey, &m.TxFeeRebateRatio, validateTxFeeRebateRatio),
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

	return nil
}

// String implements the fmt.Stringer interface.
func (m Params) String() string {
	bz, _ := yaml.Marshal(m)
	return string(bz)
}

func validateInflationRewardsRatio(v interface{}) (retErr error) {
	defer func() {
		if retErr != nil {
			retErr = fmt.Errorf("inflationRewardsRatio param: %w", retErr)
		}
	}()

	p, ok := v.(sdk.Dec)
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

	p, ok := v.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	return validateRatio(p)
}

// validateRatio is a generic ratio coefficient validator.
func validateRatio(v sdk.Dec) error {
	if v.IsNegative() {
		return fmt.Errorf("must be GTE 0.0")
	}
	if v.GTE(sdk.OneDec()) {
		return fmt.Errorf("must be LT 1.0")
	}

	return nil
}
