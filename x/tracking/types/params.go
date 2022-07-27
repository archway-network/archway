package types

import (
	"fmt"
	paramTypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"sigs.k8s.io/yaml"
)

var (
	GasTrackingEnabledParamKey = []byte("GasTrackingEnabled")
)

var _ paramTypes.ParamSet = (*Params)(nil)

// ParamKeyTable creates a new params table.
func ParamKeyTable() paramTypes.KeyTable {
	return paramTypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance.
func NewParams(gasTrackingEnabled bool) Params {
	return Params{
		GasTrackingEnabled: gasTrackingEnabled,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(true)
}

// ParamSetPairs Implements the paramTypes.ParamSet interface.
func (m *Params) ParamSetPairs() paramTypes.ParamSetPairs {
	return paramTypes.ParamSetPairs{
		paramTypes.NewParamSetPair(GasTrackingEnabledParamKey, &m.GasTrackingEnabled, validateGasTrackingEnabled),
	}
}

// Validate perform object fields validation.
func (m Params) Validate() error {
	if err := validateGasTrackingEnabled(m.GasTrackingEnabled); err != nil {
		return err
	}

	return nil
}

// String implements the fmt.Stringer interface.
func (m Params) String() string {
	bz, _ := yaml.Marshal(m)
	return string(bz)
}

func validateGasTrackingEnabled(v interface{}) (retErr error) {
	defer func() {
		if retErr != nil {
			retErr = fmt.Errorf("gasTrackingEnabled param: %w", retErr)
		}
	}()

	v, ok := v.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	return
}
