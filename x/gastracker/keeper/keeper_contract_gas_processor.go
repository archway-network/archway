package keeper

import (
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/archway-network/archway/x/gastracker"
	"github.com/cosmos/cosmos-sdk/types"
)

var _ wasmTypes.ContractGasProcessor = &Keeper{}

func (k Keeper) IngestGasRecord(ctx types.Context, records []wasmTypes.ContractGasRecord) error {
	if !k.GetParams(ctx).GasTrackingSwitch {
		return nil
	}

	for _, record := range records {
		contractAddress, err := types.AccAddressFromBech32(record.ContractAddress)
		if err != nil {
			return err
		}

		var contractMetadataExists bool
		_, err = k.GetContractMetadata(ctx, contractAddress)
		switch err {
		case gastracker.ErrContractInstanceMetadataNotFound:
			contractMetadataExists = false
		case nil:
			contractMetadataExists = true
		default:
			return err
		}

		if !contractMetadataExists {
			continue
		}

		var operation gastracker.ContractOperation
		switch record.OperationId {
		case wasmTypes.ContractOperationQuery:
			operation = gastracker.ContractOperation_CONTRACT_OPERATION_QUERY
		case wasmTypes.ContractOperationInstantiate:
			operation = gastracker.ContractOperation_CONTRACT_OPERATION_INSTANTIATION
		case wasmTypes.ContractOperationExecute:
			operation = gastracker.ContractOperation_CONTRACT_OPERATION_EXECUTION
		case wasmTypes.ContractOperationMigrate:
			operation = gastracker.ContractOperation_CONTRACT_OPERATION_MIGRATE
		case wasmTypes.ContractOperationSudo:
			operation = gastracker.ContractOperation_CONTRACT_OPERATION_SUDO
		case wasmTypes.ContractOperationReply:
			operation = gastracker.ContractOperation_CONTRACT_OPERATION_REPLY
		case wasmTypes.ContractOperationIbcPacketTimeout:
			fallthrough
		case wasmTypes.ContractOperationIbcPacketAck:
			fallthrough
		case wasmTypes.ContractOperationIbcPacketReceive:
			fallthrough
		case wasmTypes.ContractOperationIbcChannelClose:
			fallthrough
		case wasmTypes.ContractOperationIbcChannelOpen:
			fallthrough
		case wasmTypes.ContractOperationIbcChannelConnect:
			operation = gastracker.ContractOperation_CONTRACT_OPERATION_IBC
		default:
			operation = gastracker.ContractOperation_CONTRACT_OPERATION_UNSPECIFIED
		}

		k.TrackContractGasUsage(ctx, contractAddress, wasmTypes.GasConsumptionInfo{
			SDKGas: record.OriginalGas.SDKGas,
			VMGas:  k.wasmGasRegister.FromWasmVMGas(record.OriginalGas.VMGas),
		}, operation)
	}

	return nil
}

func (k Keeper) GetGasCalculationFn(ctx types.Context, contractAddress string) (func(operationId uint64, gasInfo wasmTypes.GasConsumptionInfo) wasmTypes.GasConsumptionInfo, error) {
	var contractMetadataExists bool

	passthroughFn := func(operationId uint64, gasConsumptionInfo wasmTypes.GasConsumptionInfo) wasmTypes.GasConsumptionInfo {
		return gasConsumptionInfo
	}

	doNotUse := func(operationId uint64, _ wasmTypes.GasConsumptionInfo) wasmTypes.GasConsumptionInfo {
		panic("do not use this function")
	}

	contractAddr, err := types.AccAddressFromBech32(contractAddress)
	if err != nil {
		return doNotUse, err
	}

	contractMetadata, err := k.GetContractMetadata(ctx, contractAddr)
	switch err {
	case gastracker.ErrContractInstanceMetadataNotFound:
		contractMetadataExists = false
	case nil:
		contractMetadataExists = true
	default:
		return doNotUse, err
	}

	if !contractMetadataExists {
		return passthroughFn, nil
	}

	// We are pre-fetching the configuration so that
	// gas usage is similar across all conditions.
	params := k.GetParams(ctx)
	isGasRebateToUserEnabled := params.GasRebateToUserSwitch
	isContractPremiumEnabled := params.ContractPremiumSwitch
	isGasTrackingEnabled := params.GasTrackingSwitch

	return func(operationId uint64, gasConsumptionInfo wasmTypes.GasConsumptionInfo) wasmTypes.GasConsumptionInfo {
		if !isGasTrackingEnabled {
			return gasConsumptionInfo
		}

		if isGasRebateToUserEnabled && contractMetadata.GasRebateToUser {
			updatedGas := wasmTypes.GasConsumptionInfo{
				SDKGas: (gasConsumptionInfo.SDKGas * 50) / 100,
				VMGas:  (gasConsumptionInfo.VMGas * 50) / 100,
			}
			return updatedGas
		} else if isContractPremiumEnabled && contractMetadata.CollectPremium {
			updatedGas := wasmTypes.GasConsumptionInfo{
				SDKGas: gasConsumptionInfo.SDKGas + (gasConsumptionInfo.SDKGas*contractMetadata.PremiumPercentageCharged)/100,
				VMGas:  gasConsumptionInfo.VMGas + (gasConsumptionInfo.VMGas*contractMetadata.PremiumPercentageCharged)/100,
			}
			return updatedGas
		} else {
			return gasConsumptionInfo
		}
	}, nil
}

func (k Keeper) CalculateUpdatedGas(ctx types.Context, record wasmTypes.ContractGasRecord) (wasmTypes.GasConsumptionInfo, error) {
	gasCalcFn, err := k.GetGasCalculationFn(ctx, record.ContractAddress)
	if err != nil {
		return wasmTypes.GasConsumptionInfo{}, nil
	}

	return gasCalcFn(record.OperationId, record.OriginalGas), nil
}
