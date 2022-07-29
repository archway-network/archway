package keeper_test

import (
	"testing"

	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/stretchr/testify/suite"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/tracking/types"
)

type KeeperTestSuite struct {
	suite.Suite

	chain *e2eTesting.TestChain
}

func (s *KeeperTestSuite) SetupTest() {
	s.chain = e2eTesting.NewTestChain(s.T(), 1)
}

func TestTrackingKeeper(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func ContractOperationToWASM(opType types.ContractOperation) uint64 {
	switch opType {
	case types.ContractOperation_CONTRACT_OPERATION_INSTANTIATION:
		return wasmTypes.ContractOperationInstantiate
	case types.ContractOperation_CONTRACT_OPERATION_EXECUTION:
		return wasmTypes.ContractOperationExecute
	case types.ContractOperation_CONTRACT_OPERATION_QUERY:
		return wasmTypes.ContractOperationQuery
	case types.ContractOperation_CONTRACT_OPERATION_MIGRATE:
		return wasmTypes.ContractOperationMigrate
	case types.ContractOperation_CONTRACT_OPERATION_IBC:
		return wasmTypes.ContractOperationIbcPacketReceive
	case types.ContractOperation_CONTRACT_OPERATION_SUDO:
		return wasmTypes.ContractOperationSudo
	case types.ContractOperation_CONTRACT_OPERATION_REPLY:
		return wasmTypes.ContractOperationReply
	case types.ContractOperation_CONTRACT_OPERATION_UNSPECIFIED:
		fallthrough
	default:
		return wasmTypes.ContractOperationUnknown
	}
}
