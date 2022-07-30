package keeper_test

import (
	"math/rand"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

// mockContractViewer mocks x/wasmd module dependency.
type mockContractViewer struct {
	contractAdminSet map[string]string // key: contractAddr, value: adminAddr
}

func newMockContractViewer() *mockContractViewer {
	return &mockContractViewer{
		contractAdminSet: make(map[string]string),
	}
}

// AddContractAdmin adds a contract admin link.
func (v *mockContractViewer) AddContractAdmin(contractAddr, adminAddr string) {
	v.contractAdminSet[contractAddr] = adminAddr
}

func (v mockContractViewer) GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmdTypes.ContractInfo {
	adminAddr, found := v.contractAdminSet[contractAddress.String()]
	if !found {
		return nil
	}

	return &wasmdTypes.ContractInfo{
		Admin: adminAddr,
	}
}

// GetRandomContractOperationType returns a random wasmd contract operation type.
func GetRandomContractOperationType() uint64 {
	idx := rand.Intn(len(allContractOperationTypes))
	return allContractOperationTypes[idx]
}
