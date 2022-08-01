package testutils

import (
	"math/rand"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"

	trackingTypes "github.com/archway-network/archway/x/tracking/types"
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

// RewardsContractOperationToWASM converts x/tracking contract operation to wasmd type.
func RewardsContractOperationToWASM(opType trackingTypes.ContractOperation) uint64 {
	switch opType {
	case trackingTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION:
		return wasmdTypes.ContractOperationInstantiate
	case trackingTypes.ContractOperation_CONTRACT_OPERATION_EXECUTION:
		return wasmdTypes.ContractOperationExecute
	case trackingTypes.ContractOperation_CONTRACT_OPERATION_QUERY:
		return wasmdTypes.ContractOperationQuery
	case trackingTypes.ContractOperation_CONTRACT_OPERATION_MIGRATE:
		return wasmdTypes.ContractOperationMigrate
	case trackingTypes.ContractOperation_CONTRACT_OPERATION_IBC:
		return wasmdTypes.ContractOperationIbcPacketReceive
	case trackingTypes.ContractOperation_CONTRACT_OPERATION_SUDO:
		return wasmdTypes.ContractOperationSudo
	case trackingTypes.ContractOperation_CONTRACT_OPERATION_REPLY:
		return wasmdTypes.ContractOperationReply
	case trackingTypes.ContractOperation_CONTRACT_OPERATION_UNSPECIFIED:
		fallthrough
	default:
		return wasmdTypes.ContractOperationUnknown
	}
}

// WASMContractOperationToRewards converts wasmd operation type to x/tracking type.
func WASMContractOperationToRewards(opType uint64) trackingTypes.ContractOperation {
	switch opType {
	case wasmdTypes.ContractOperationQuery:
		return trackingTypes.ContractOperation_CONTRACT_OPERATION_QUERY
	case wasmdTypes.ContractOperationInstantiate:
		return trackingTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION
	case wasmdTypes.ContractOperationExecute:
		return trackingTypes.ContractOperation_CONTRACT_OPERATION_EXECUTION
	case wasmdTypes.ContractOperationMigrate:
		return trackingTypes.ContractOperation_CONTRACT_OPERATION_MIGRATE
	case wasmdTypes.ContractOperationSudo:
		return trackingTypes.ContractOperation_CONTRACT_OPERATION_SUDO
	case wasmdTypes.ContractOperationReply:
		return trackingTypes.ContractOperation_CONTRACT_OPERATION_REPLY
	case wasmdTypes.ContractOperationIbcPacketTimeout:
		fallthrough
	case wasmdTypes.ContractOperationIbcPacketAck:
		fallthrough
	case wasmdTypes.ContractOperationIbcPacketReceive:
		fallthrough
	case wasmdTypes.ContractOperationIbcChannelClose:
		fallthrough
	case wasmdTypes.ContractOperationIbcChannelOpen:
		fallthrough
	case wasmdTypes.ContractOperationIbcChannelConnect:
		return trackingTypes.ContractOperation_CONTRACT_OPERATION_IBC
	default:
		return trackingTypes.ContractOperation_CONTRACT_OPERATION_UNSPECIFIED
	}
}
