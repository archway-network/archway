package gastracker

import (
	fmt "fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"

	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

const (
	DefaultParamSpace = ModuleName

	DefaultMaxGasForLocalFeeGrant  = 1000000
	DefaultMaxGasForGlobalFeeGrant = 3000000

	DefaultGasTrackingSwitch      = true
	GasDappInflationRewardsSwitch = true
	DefaultGasRebateSwitch        = true
	DefaultGasRebateToUserSwitch  = true
	DefaultContractPremiumSwitch  = true
)

var (
	ParamsKeyGasTrackingSwitch     = []byte("GasTrackingSwitch")
	ParamsKeyDappInflationRewards  = []byte("DappInflationRewards")
	ParamsKeyGasRebateSwitch       = []byte("GasRebateSwitch")
	ParamsKeyGasRebateToUserSwitch = []byte("GasRebateToUserSwitch")
	ParamsKeyContractPremiumSwitch = []byte("ContractPremiumSwitch")

	ParamsKeyMaxGasForContractFeeGrant = []byte("MaxGasForContractFeeGrant")
	ParamsKeyMaxGasForGlobalGrant      = []byte("MaxGasForGlobalFeeGrant")
)

type Params struct {
	GasTrackingSwitch             bool
	GasDappInflationRewardsSwitch bool
	GasRebateSwitch               bool
	GasRebateToUserSwitch         bool
	ContractPremiumSwitch         bool

	MaxGasForContractFeeGrant uint64
	MaxGasForGlobalFeeGrant   uint64
}

var _ paramstypes.ParamSet = &Params{}

func DefaultParams(ctx sdk.Context) Params {
	defaultParams := Params{
		GasTrackingSwitch:             DefaultGasTrackingSwitch,
		GasDappInflationRewardsSwitch: GasDappInflationRewardsSwitch,
		GasRebateSwitch:               DefaultGasRebateSwitch,
		GasRebateToUserSwitch:         DefaultGasRebateToUserSwitch,
		ContractPremiumSwitch:         DefaultContractPremiumSwitch,
	}

	if ctx.BlockGasMeter().Limit() == 0 {
		defaultParams.MaxGasForGlobalFeeGrant = DefaultMaxGasForGlobalFeeGrant
		defaultParams.MaxGasForContractFeeGrant = DefaultMaxGasForLocalFeeGrant
	} else {
		defaultParams.MaxGasForGlobalFeeGrant = (ctx.BlockGasMeter().Limit() * 40) / 100
		defaultParams.MaxGasForContractFeeGrant = (ctx.BlockGasMeter().Limit() * 5) / 100
	}

	return defaultParams
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
		paramstypes.NewParamSetPair(ParamsKeyMaxGasForGlobalGrant, &p.MaxGasForGlobalFeeGrant, validateUint64),
		paramstypes.NewParamSetPair(ParamsKeyMaxGasForContractFeeGrant, &p.MaxGasForContractFeeGrant, validateUint64),
	}
}

func validateSwitch(i interface{}) error {
	if _, ok := i.(bool); !ok {
		return fmt.Errorf("Invalid parameter type %T", i)
	}
	return nil
}

func validateUint64(i interface{}) error {
	if _, ok := i.(uint64); !ok {
		return fmt.Errorf("invalid parameter type %T", i)
	}
	return nil
}
