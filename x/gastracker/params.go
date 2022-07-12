package gastracker

import (
	fmt "fmt"

	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

const (
	DefaultParamSpace = ModuleName
)

var (
	ParamsKeyGasTrackingSwitch     = []byte("GasTrackingSwitch")
	ParamsKeyDappInflationRewards  = []byte("DappInflationRewards")
	ParamsKeyGasRebateSwitch       = []byte("GasRebateSwitch")
	ParamsKeyGasRebateToUserSwitch = []byte("GasRebateToUserSwitch")
	ParamsKeyContractPremiumSwitch = []byte("ContractPremiumSwitch")

	ParamsKeyInflationRewardQuotaPercentage = []byte("InflationRewardQuotaPercentage")
	ParamsKeyGasRebatePercentage            = []byte("GasRebatePercentage")
)

type Params struct {
	GasTrackingSwitch             bool
	GasDappInflationRewardsSwitch bool
	GasRebateSwitch               bool
	GasRebateToUserSwitch         bool
	ContractPremiumSwitch         bool

	InflationRewardQuotaPercentage uint64
	GasRebatePercentage            uint64
}

var (
	DefaultGasTrackingSwitch      = true
	GasDappInflationRewardsSwitch = true
	DefaultGasRebateSwitch        = true
	DefaultGasRebateToUserSwitch  = true
	DefaultContractPremiumSwitch  = true

	DefaultInflationRewardQuotaPercentage uint64 = 20
	DefaultGasRebatePercentage            uint64 = 50
)

var _ paramstypes.ParamSet = &Params{}

func DefaultParams() Params {
	return Params{
		GasTrackingSwitch:             DefaultGasTrackingSwitch,
		GasDappInflationRewardsSwitch: GasDappInflationRewardsSwitch,
		GasRebateSwitch:               DefaultGasRebateSwitch,
		GasRebateToUserSwitch:         DefaultGasRebateToUserSwitch,
		ContractPremiumSwitch:         DefaultContractPremiumSwitch,

		InflationRewardQuotaPercentage: DefaultInflationRewardQuotaPercentage,
		GasRebatePercentage:            DefaultGasRebatePercentage,
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
		paramstypes.NewParamSetPair(ParamsKeyGasTrackingSwitch, &p.GasTrackingSwitch, validateSwitch),
		paramstypes.NewParamSetPair(ParamsKeyDappInflationRewards, &p.GasDappInflationRewardsSwitch, validateSwitch),
		paramstypes.NewParamSetPair(ParamsKeyGasRebateSwitch, &p.GasRebateSwitch, validateSwitch),
		paramstypes.NewParamSetPair(ParamsKeyGasRebateToUserSwitch, &p.GasRebateToUserSwitch, validateSwitch),
		paramstypes.NewParamSetPair(ParamsKeyContractPremiumSwitch, &p.ContractPremiumSwitch, validateSwitch),

		paramstypes.NewParamSetPair(ParamsKeyInflationRewardQuotaPercentage, &p.InflationRewardQuotaPercentage, validateUint64Percentage),
		paramstypes.NewParamSetPair(ParamsKeyGasRebatePercentage, &p.GasRebatePercentage, validateUint64Percentage),
	}
}

func validateSwitch(i interface{}) error {
	if _, ok := i.(bool); !ok {
		return fmt.Errorf("Invalid parameter type %T", i)
	}
	return nil
}

func validateUint64Percentage(i interface{}) error {
	if val, ok := i.(uint64); !ok {
		return fmt.Errorf("invalid parameter type %T", i)
	} else {
		if val > 100 {
			return fmt.Errorf("percentage cannot be greater than 100, found: %d", val)
		}
	}
	return nil
}
