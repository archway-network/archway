package types

import (
	fmt "fmt"

	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

const (
	DefaultParamSpace = ModuleName
)

var (
	KeyGasTrackingSwitch     = []byte("GasTrackingSwitch")
	KeyGasRebateSwitch       = []byte("GasRebateSwitch")
	KeyGasRebateToUserSwitch = []byte("GasRebateToUserSwitch")
	KeyContractPremiumSwitch = []byte("ContractPremiumSwitch")
)

type Params struct {
	GasTrackingSwitch     bool
	GasRebateSwitch       bool
	GasRebateToUserSwitch bool
	ContractPremiumSwitch bool
}

var (
	DefaultGasTrackingSwitch     = true
	DefaultGasRebateSwitch       = true
	DefaultGasRebateToUserSwitch = true
	DefaultContractPremiumSwitch = true
)

var _ paramstypes.ParamSet = &Params{}

func DefaultParams() Params {
	return Params{
		GasTrackingSwitch:     DefaultGasTrackingSwitch,
		GasRebateSwitch:       DefaultGasRebateSwitch,
		GasRebateToUserSwitch: DefaultGasRebateToUserSwitch,
		ContractPremiumSwitch: DefaultContractPremiumSwitch,
	}
}

func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyGasTrackingSwitch, &p.GasTrackingSwitch, validateSwitch),
		paramstypes.NewParamSetPair(KeyGasRebateSwitch, &p.GasRebateSwitch, validateSwitch),
		paramstypes.NewParamSetPair(KeyGasRebateToUserSwitch, &p.GasRebateToUserSwitch, validateSwitch),
		paramstypes.NewParamSetPair(KeyContractPremiumSwitch, &p.ContractPremiumSwitch, validateSwitch),
	}
}

func validateSwitch(i interface{}) error {
	if _, ok := i.(bool); !ok {
		return fmt.Errorf("Invalid parameter type %T", i)
	}
	return nil
}
