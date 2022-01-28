package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/archway-network/archway/x/gastracker/types"
)

func ParamChanges(r *rand.Rand, cdc codec.Codec) []simtypes.ParamChange {
	params := RandomParams(r)
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.ParamsKeyContractPremiumSwitch),
			func(r *rand.Rand) string {
				return fmt.Sprintf(`"%v"`, params.ContractPremiumSwitch)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.ParamsKeyGasRebateToUserSwitch),
			func(r *rand.Rand) string {
				return fmt.Sprintf(`"%v"`, params.GasRebateSwitch)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.ParamsKeyDappInflationRewards),
			func(r *rand.Rand) string {
				return fmt.Sprintf(`"%v"`, params.GasDappInflationRewardsSwitch)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.ParamsKeyGasRebateSwitch),
			func(r *rand.Rand) string {
				return fmt.Sprintf(`"%v"`, params.GasRebateSwitch)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.ParamsKeyGasTrackingSwitch),
			func(r *rand.Rand) string {
				return fmt.Sprintf(`"%v"`, params.GasTrackingSwitch)
			},
		),
	}
}

func RandomParams(r *rand.Rand) types.Params {
	return types.Params{
		GasTrackingSwitch:             simtypes.RandIntBetween(r, 1, 50)%2 == 0,
		GasDappInflationRewardsSwitch: simtypes.RandIntBetween(r, 3, 52)%2 == 0,
		GasRebateSwitch:               simtypes.RandIntBetween(r, 5, 54)%2 == 0,
		GasRebateToUserSwitch:         simtypes.RandIntBetween(r, 7, 56)%2 == 0,
		ContractPremiumSwitch:         simtypes.RandIntBetween(r, 9, 58)%2 == 0,
	}
}
