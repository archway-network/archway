package testutils

import (
	wasmKeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmVmTypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ wasmKeeper.Messenger = (*MockMessenger)(nil)

// MockContractViewer mocks x/wasmd module dependency.
// Mock returns a contract info if admin is set.
type MockContractViewer struct {
	contractAdminSet map[string]string // key: contractAddr, value: adminAddr
}

// NewMockContractViewer creates a new MockContractViewer instance.
func NewMockContractViewer() *MockContractViewer {
	return &MockContractViewer{
		contractAdminSet: make(map[string]string),
	}
}

// AddContractAdmin adds a contract admin link.
func (v *MockContractViewer) AddContractAdmin(contractAddr, adminAddr string) {
	v.contractAdminSet[contractAddr] = adminAddr
}

// GetContractInfo returns a contract info if admin is set.
func (v MockContractViewer) GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmdTypes.ContractInfo {
	adminAddr, found := v.contractAdminSet[contractAddress.String()]
	if !found {
		return nil
	}

	return &wasmdTypes.ContractInfo{
		Admin: adminAddr,
	}
}

// MockMessenger mocks x/wasmd module dependency.
type MockMessenger struct{}

// NewMockMessenger creates a new MockMessenger instance.
func NewMockMessenger() *MockMessenger {
	return &MockMessenger{}
}

// DispatchMsg implements the wasmKeeper.Messenger interface.
func (m MockMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmVmTypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
	return nil, nil, nil
}
