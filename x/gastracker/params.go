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
	ParamsKeyGasTrackingSwitch              = []byte("GasTrackingSwitch")
	ParamsKeyDappInflationRewardsSwitch     = []byte("DefaultDappInflationRewardsSwitch")
	ParamsKeyGasRebateSwitch                = []byte("GasRebateSwitch")
	ParamsKeyGasRebateToUserSwitch          = []byte("GasRebateToUserSwitch")
	ParamsKeyContractPremiumSwitch          = []byte("ContractPremiumSwitch")
	ParamsKeyInflationRewardQuotaPercentage = []byte("InflationRewardQuotaPercentage")
	ParamsKeyInflationRewardCapSwitch       = []byte("InflationRewardCapSwitch")
	ParamsKeyInflationRewardCapPercentage   = []byte("InflationRewardCapPercentage")
)

type Params struct {
	GasTrackingSwitch          bool
	DappInflationRewardsSwitch bool
	GasRebateSwitch            bool
	GasRebateToUserSwitch      bool
	ContractPremiumSwitch      bool

	InflationRewardQuotaPercentage uint64
	InflationRewardCapSwitch       bool
	InflationRewardCapPercentage   uint64
}

const (
	DefaultGasTrackingSwitch          = true
	DefaultDappInflationRewardsSwitch = true
	DefaultGasRebateSwitch            = true
	DefaultGasRebateToUserSwitch      = true
	DefaultContractPremiumSwitch      = true

	DefaultInflationRewardQuotaPercentage uint64 = 20
	DefaultInflationRewardCapSwitch              = false
	DefaultInflationRewardCapPercentage   uint64 = 100
)

var _ paramstypes.ParamSet = &Params{}

func DefaultParams() Params {
	return Params{
		GasTrackingSwitch:          DefaultGasTrackingSwitch,
		DappInflationRewardsSwitch: DefaultDappInflationRewardsSwitch,
		GasRebateSwitch:            DefaultGasRebateSwitch,
		GasRebateToUserSwitch:      DefaultGasRebateToUserSwitch,
		ContractPremiumSwitch:      DefaultContractPremiumSwitch,

		InflationRewardQuotaPercentage: DefaultInflationRewardQuotaPercentage,
		InflationRewardCapSwitch:       DefaultInflationRewardCapSwitch,
		InflationRewardCapPercentage:   DefaultInflationRewardCapPercentage,
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
		paramstypes.NewParamSetPair(ParamsKeyDappInflationRewardsSwitch, &p.DappInflationRewardsSwitch, validateSwitch),
		paramstypes.NewParamSetPair(ParamsKeyGasRebateSwitch, &p.GasRebateSwitch, validateSwitch),
		paramstypes.NewParamSetPair(ParamsKeyGasRebateToUserSwitch, &p.GasRebateToUserSwitch, validateSwitch),
		paramstypes.NewParamSetPair(ParamsKeyContractPremiumSwitch, &p.ContractPremiumSwitch, validateSwitch),

		paramstypes.NewParamSetPair(ParamsKeyInflationRewardQuotaPercentage, &p.InflationRewardQuotaPercentage, validateUint64Percentage),
		paramstypes.NewParamSetPair(ParamsKeyInflationRewardCapSwitch, &p.InflationRewardCapSwitch, validateSwitch),
		paramstypes.NewParamSetPair(ParamsKeyInflationRewardCapPercentage, &p.InflationRewardCapPercentage, validateUint64Percentage),
	}
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

func validateSwitch(i interface{}) error {
	if _, ok := i.(bool); !ok {
		return fmt.Errorf("invalid parameter type %T", i)
	}
	return nil
}
