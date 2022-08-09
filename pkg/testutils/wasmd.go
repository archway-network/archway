package testutils

import (
	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
