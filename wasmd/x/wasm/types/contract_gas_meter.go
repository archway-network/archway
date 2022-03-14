package types

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.GasMeter = &ContractSDKGasMeter{}

type ContractSDKGasMeter struct {
	actualGasConsumed          sdk.Gas
	requestedGas               sdk.Gas
	underlyingGasMeter         sdk.GasMeter
	contractAddress            string
	contractGasCalculationFunc func(operationId uint64, info GasConsumptionInfo) GasConsumptionInfo
}

func NewContractGasMeter(underlyingGasMeter sdk.GasMeter, gasCalculationFunc func(uint64, GasConsumptionInfo) GasConsumptionInfo, contractAddress string) ContractSDKGasMeter {
	return ContractSDKGasMeter{
		actualGasConsumed:          0,
		requestedGas:               0,
		contractGasCalculationFunc: gasCalculationFunc,
		underlyingGasMeter:         underlyingGasMeter,
		contractAddress:            contractAddress,
	}
}

func (c *ContractSDKGasMeter) GetContractAddress() string {
	return c.contractAddress
}

func (c *ContractSDKGasMeter) GetGasStat() (sdk.Gas, sdk.Gas) {
	return c.requestedGas, c.actualGasConsumed
}

func (c *ContractSDKGasMeter) GasConsumed() storetypes.Gas {
	return c.underlyingGasMeter.GasConsumed()
}

func (c *ContractSDKGasMeter) GasConsumedToLimit() storetypes.Gas {
	return c.underlyingGasMeter.GasConsumedToLimit()
}

func (c *ContractSDKGasMeter) Limit() storetypes.Gas {
	return c.underlyingGasMeter.Limit()
}

func (c *ContractSDKGasMeter) ConsumeGas(amount storetypes.Gas, descriptor string) {
	updatedGasInfo := c.contractGasCalculationFunc(ContractOperationUnknown, GasConsumptionInfo{SDKGas: amount})
	c.underlyingGasMeter.ConsumeGas(updatedGasInfo.SDKGas, descriptor)
	c.requestedGas += amount
	c.actualGasConsumed += updatedGasInfo.SDKGas
}

func (c *ContractSDKGasMeter) RefundGas(amount storetypes.Gas, descriptor string) {
	updatedGasInfo := c.contractGasCalculationFunc(ContractOperationUnknown, GasConsumptionInfo{SDKGas: amount})
	c.underlyingGasMeter.RefundGas(updatedGasInfo.SDKGas, descriptor)
	c.requestedGas -= amount
	c.actualGasConsumed -= updatedGasInfo.SDKGas
}

func (c *ContractSDKGasMeter) IsPastLimit() bool {
	return c.underlyingGasMeter.IsPastLimit()
}

func (c *ContractSDKGasMeter) IsOutOfGas() bool {
	return c.underlyingGasMeter.IsOutOfGas()
}

func (c *ContractSDKGasMeter) String() string {
	return c.underlyingGasMeter.String()
}
