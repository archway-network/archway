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
	KeyDappInflationRewards  = []byte("DappInflationRewards")
	KeyGasRebateSwitch       = []byte("GasRebateSwitch")
	KeyGasRebateToUserSwitch = []byte("GasRebateToUserSwitch")
	KeyContractPremiumSwitch = []byte("ContractPremiumSwitch")
)

type Params struct {
	GasTrackingSwitch             bool
	GasDappInflationRewardsSwitch bool
	GasRebateSwitch               bool
	GasRebateToUserSwitch         bool
	ContractPremiumSwitch         bool
}

var (
	DefaultGasTrackingSwitch      = true
	GasDappInflationRewardsSwitch = true
	DefaultGasRebateSwitch        = true
	DefaultGasRebateToUserSwitch  = true
	DefaultContractPremiumSwitch  = true
)

var _ paramstypes.ParamSet = &Params{}

func DefaultParams() Params {
	return Params{
		GasTrackingSwitch:             DefaultGasTrackingSwitch,
		GasDappInflationRewardsSwitch: GasDappInflationRewardsSwitch,
		GasRebateSwitch:               DefaultGasRebateSwitch,
		GasRebateToUserSwitch:         DefaultGasRebateToUserSwitch,
		ContractPremiumSwitch:         DefaultContractPremiumSwitch,
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
		paramstypes.NewParamSetPair(KeyDappInflationRewards, &p.GasDappInflationRewardsSwitch, validateSwitch),
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
