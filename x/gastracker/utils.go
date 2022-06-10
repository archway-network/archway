package gastracker

import (
	"github.com/CosmWasm/wasmd/x/wasm/types"
	"math/big"
)

// This function is mainly used to safely convert the uint64 number to big.Int and
// then to sdk.Dec/sdk.Int
func ConvertUint64ToBigInt(n uint64) *big.Int {
	return big.NewInt(0).SetUint64(n)
}

func AddPremiumGasInConsumption(metadata ContractInstanceMetadata, gasConsumptionInfo types.GasConsumptionInfo) types.GasConsumptionInfo {
	return types.GasConsumptionInfo{
		SDKGas: gasConsumptionInfo.SDKGas + (gasConsumptionInfo.SDKGas*metadata.PremiumPercentageCharged)/100,
		VMGas:  gasConsumptionInfo.VMGas + (gasConsumptionInfo.VMGas*metadata.PremiumPercentageCharged)/100,
	}
}

func DeductGasRebateFromConsumption(metadata ContractInstanceMetadata, gasConsumptionInfo types.GasConsumptionInfo, percentage uint64) types.GasConsumptionInfo {
	return types.GasConsumptionInfo{
		SDKGas: (gasConsumptionInfo.SDKGas * percentage) / 100,
		VMGas:  (gasConsumptionInfo.VMGas * percentage) / 100,
	}
}
