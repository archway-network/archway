package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	ContractOperationInstantiate uint64 = iota
	ContractOperationExecute
	ContractOperationQuery
	ContractOperationMigrate
	ContractOperationSudo
	ContractOperationReply
	ContractOperationIbcChannelOpen
	ContractOperationIbcChannelConnect
	ContractOperationIbcChannelClose
	ContractOperationIbcPacketReceive
	ContractOperationIbcPacketAck
	ContractOperationIbcPacketTimeout
)

type ContractGasRecord struct {
	OperationId uint64
	ContractAddress string
	GasConsumed uint64
}

type ContractGasProcessor interface {
	IngestGasRecord(ctx sdk.Context, records []ContractGasRecord) error
	CalculateUpdatedGas(ctx sdk.Context, record ContractGasRecord) (uint64, error)
}

// NoOpContractGasProcessor is no-op gas processor
type NoOpContractGasProcessor struct {

}

func (n *NoOpContractGasProcessor) IngestGasRecord(_ sdk.Context, _ []ContractGasRecord) error {
	return nil
}

func (n *NoOpContractGasProcessor) CalculateUpdatedGas(_ sdk.Context, _ ContractGasRecord) (uint64, error) {
	return 0, nil
}
