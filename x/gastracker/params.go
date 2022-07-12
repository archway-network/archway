package gastracker

import (
	fmt "fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	DefaultParamSpace = ModuleName
)

var (
	ParamsKeyGasTrackingSwitch         = []byte("GasTrackingSwitch")
	ParamsKeyDappInflationRewards      = []byte("DappInflationRewards")
	ParamsKeyGasRebateSwitch           = []byte("GasRebateSwitch")
	ParamsKeyGasRebateToUserSwitch     = []byte("GasRebateToUserSwitch")
	ParamsKeyContractPremiumSwitch     = []byte("ContractPremiumSwitch")
	ParamsKeyDappInflationRewardsRatio = []byte("ParamsKeyDappInflationRewardsRatio")
	ParamsKeyDappTxFeeRebateRatio      = []byte("ParamsKeyDappTxFeeRebateRatio")
)

var _ paramstypes.ParamSet = &Params{}

func DefaultParams() Params {
	return Params{
		GasTrackingSwitch:             true,
		GasDappInflationRewardsSwitch: true,
		GasRebateSwitch:               true,
		GasRebateToUserSwitch:         true,
		ContractPremiumSwitch:         true,
		DappInflationRewardsRatio:     sdk.MustNewDecFromStr("0.25"), // 25%
		DappTxFeeRebateRatio:          sdk.MustNewDecFromStr("0.5"),  // 50%
	}
}

func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

func (m *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(ParamsKeyGasTrackingSwitch, &m.GasTrackingSwitch, validateSwitch),
		paramstypes.NewParamSetPair(ParamsKeyDappInflationRewards, &m.GasDappInflationRewardsSwitch, validateSwitch),
		paramstypes.NewParamSetPair(ParamsKeyGasRebateSwitch, &m.GasRebateSwitch, validateSwitch),
		paramstypes.NewParamSetPair(ParamsKeyGasRebateToUserSwitch, &m.GasRebateToUserSwitch, validateSwitch),
		paramstypes.NewParamSetPair(ParamsKeyContractPremiumSwitch, &m.ContractPremiumSwitch, validateSwitch),
		paramstypes.NewParamSetPair(ParamsKeyDappInflationRewardsRatio, &m.DappInflationRewardsRatio, func(value interface{}) error {
			ratio, ok := value.(sdk.Dec)
			if !ok {
				return fmt.Errorf("invalid param field type in %s.dapp_inflation_rewards_ratio: %T, expected sdk.Dec", proto.MessageName(m), value)
			}

			if ratio.LT(sdk.ZeroDec()) || ratio.GT(sdk.OneDec()) {
				return fmt.Errorf("invalid param value in %s.dapp_inflation_rewards_ratio, 0 <= value <= 1, got: %s", proto.MessageName(m), ratio)
			}

			return nil
		}),
		paramstypes.NewParamSetPair(ParamsKeyDappTxFeeRebateRatio, &m.DappTxFeeRebateRatio, func(value interface{}) error {
			ratio, ok := value.(sdk.Dec)
			if !ok {
				return fmt.Errorf("invalid param field type in %s.dapp_tx_fee_rebate_ratio: %T, expected sdk.Dec", proto.MessageName(m), value)
			}

			if ratio.LT(sdk.ZeroDec()) || ratio.GTE(sdk.OneDec()) {
				return fmt.Errorf("invalid param value in %s.dapp_tx_fee_rebate_ratio, 0 <= value < 1, got: %s", proto.MessageName(m), ratio)
			}

			return nil
		}),
	}
}

func validateSwitch(i interface{}) error {
	if _, ok := i.(bool); !ok {
		return fmt.Errorf("Invalid parameter type %T", i)
	}
	return nil
}
