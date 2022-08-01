package testutils

import (
	"math/rand"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"

	treckingTypes "github.com/archway-network/archway/x/tracking/types"
)

var allContractOperationTypes = []uint64{
	wasmdTypes.ContractOperationInstantiate,
	wasmdTypes.ContractOperationExecute,
	wasmdTypes.ContractOperationQuery,
	wasmdTypes.ContractOperationMigrate,
	wasmdTypes.ContractOperationSudo,
	wasmdTypes.ContractOperationReply,
	wasmdTypes.ContractOperationIbcChannelOpen,
	wasmdTypes.ContractOperationIbcChannelConnect,
	wasmdTypes.ContractOperationIbcChannelClose,
	wasmdTypes.ContractOperationIbcPacketReceive,
	wasmdTypes.ContractOperationIbcPacketAck,
	wasmdTypes.ContractOperationIbcPacketTimeout,
	wasmdTypes.ContractOperationUnknown,
}

// GetRandomContractOperationType returns a random wasmd contract operation type.
func GetRandomContractOperationType() uint64 {
	idx := rand.Intn(len(allContractOperationTypes))
	return allContractOperationTypes[idx]
}

// ContractOperationToWASM converts x/tracking contract operation to wasmd type.
func ContractOperationToWASM(opType treckingTypes.ContractOperation) uint64 {
	switch opType {
	case treckingTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION:
		return wasmdTypes.ContractOperationInstantiate
	case treckingTypes.ContractOperation_CONTRACT_OPERATION_EXECUTION:
		return wasmdTypes.ContractOperationExecute
	case treckingTypes.ContractOperation_CONTRACT_OPERATION_QUERY:
		return wasmdTypes.ContractOperationQuery
	case treckingTypes.ContractOperation_CONTRACT_OPERATION_MIGRATE:
		return wasmdTypes.ContractOperationMigrate
	case treckingTypes.ContractOperation_CONTRACT_OPERATION_IBC:
		return wasmdTypes.ContractOperationIbcPacketReceive
	case treckingTypes.ContractOperation_CONTRACT_OPERATION_SUDO:
		return wasmdTypes.ContractOperationSudo
	case treckingTypes.ContractOperation_CONTRACT_OPERATION_REPLY:
		return wasmdTypes.ContractOperationReply
	case treckingTypes.ContractOperation_CONTRACT_OPERATION_UNSPECIFIED:
		fallthrough
	default:
		return wasmdTypes.ContractOperationUnknown
	}
}
