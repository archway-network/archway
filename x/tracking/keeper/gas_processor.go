package keeper

import (
	"fmt"

	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/tracking/types"
)

var _ wasmTypes.ContractGasProcessor = &Keeper{}

// IngestGasRecord implements the wasmTypes.ContractGasProcessor interface.
// It is called by the wasmd to track contract gas records.
func (k Keeper) IngestGasRecord(ctx sdk.Context, records []wasmTypes.ContractGasRecord) error {
	// Ingest operation for every record
	for _, record := range records {
		contractAddr, err := sdk.AccAddressFromBech32(record.ContractAddress)
		if err != nil {
			return fmt.Errorf("parsing contract address: %w", err)
		}

		var opType types.ContractOperation
		switch record.OperationId {
		case wasmTypes.ContractOperationQuery:
			opType = types.ContractOperation_CONTRACT_OPERATION_QUERY
		case wasmTypes.ContractOperationInstantiate:
			opType = types.ContractOperation_CONTRACT_OPERATION_INSTANTIATION
		case wasmTypes.ContractOperationExecute:
			opType = types.ContractOperation_CONTRACT_OPERATION_EXECUTION
		case wasmTypes.ContractOperationMigrate:
			opType = types.ContractOperation_CONTRACT_OPERATION_MIGRATE
		case wasmTypes.ContractOperationSudo:
			opType = types.ContractOperation_CONTRACT_OPERATION_SUDO
		case wasmTypes.ContractOperationReply:
			opType = types.ContractOperation_CONTRACT_OPERATION_REPLY
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
			opType = types.ContractOperation_CONTRACT_OPERATION_IBC
		default:
			opType = types.ContractOperation_CONTRACT_OPERATION_UNSPECIFIED
		}

		k.TrackNewContractOperation(
			ctx,
			contractAddr,
			opType,
			k.WasmGasRegister.FromWasmVMGas(record.OriginalGas.VMGas),
			record.OriginalGas.SDKGas,
		)
	}

	return nil
}

// GetGasCalculationFn implements the wasmTypes.ContractGasProcessor interface.
// It is called by the wasmd to get the gas consumption adjustment function for a contract.
// This is a no-op function since we don't change gas values atm.
func (k Keeper) GetGasCalculationFn(ctx sdk.Context, contractAddrBz string) (func(operationId uint64, gasInfo wasmTypes.GasConsumptionInfo) wasmTypes.GasConsumptionInfo, error) {
	return func(operationID uint64, gasConsumptionInfo wasmTypes.GasConsumptionInfo) wasmTypes.GasConsumptionInfo {
		return gasConsumptionInfo
	}, nil
}

// CalculateUpdatedGas implements the wasmTypes.ContractGasProcessor interface.
// It is called by the wasmd to modify a gas consumption record.
func (k Keeper) CalculateUpdatedGas(ctx sdk.Context, record wasmTypes.ContractGasRecord) (wasmTypes.GasConsumptionInfo, error) {
	gasCalcFn, err := k.GetGasCalculationFn(ctx, record.ContractAddress)
	if err != nil {
		return wasmTypes.GasConsumptionInfo{}, err
	}

	return gasCalcFn(record.OperationId, record.OriginalGas), nil
}
