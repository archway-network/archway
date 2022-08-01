package testutils

import (
	"math/rand"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
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
