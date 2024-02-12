package testutils

import (
	"errors"

	types0 "github.com/cosmos/cosmos-sdk/types"
	types2 "github.com/cosmos/cosmos-sdk/x/capability/types"
	types3 "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
)

// MockICAControllerKeeper is a mock of ICAControllerKeeper interface.
type MockICAControllerKeeper struct {
	returnErrForRegisterInterchainAccount bool
	interchainAccountAddress              string
}

// NewMockICAControllerKeeper creates a new mock instance.
func NewMockICAControllerKeeper() *MockICAControllerKeeper {
	return &MockICAControllerKeeper{
		returnErrForRegisterInterchainAccount: true,
	}
}

// GetActiveChannelID mocks base method.
func (m *MockICAControllerKeeper) GetActiveChannelID(ctx types0.Context, connectionID, portID string) (string, bool) {
	return "", true
}

// GetInterchainAccountAddress mocks base method.
func (m *MockICAControllerKeeper) GetInterchainAccountAddress(ctx types0.Context, connectionID, portID string) (string, bool) {
	if m.interchainAccountAddress != "" {
		return m.interchainAccountAddress, true
	}
	return "", false
}

// RegisterInterchainAccount mocks base method.
func (m *MockICAControllerKeeper) RegisterInterchainAccount(ctx types0.Context, connectionID, owner, version string) error {
	if m.returnErrForRegisterInterchainAccount {
		return errors.New("failed to create RegisterInterchainAccount")
	}
	return nil
}

// SendTx mocks base method.
func (m *MockICAControllerKeeper) SendTx(ctx types0.Context, chanCap *types2.Capability, connectionID, portID string, icaPacketData types3.InterchainAccountPacketData, timeoutTimestamp uint64) (uint64, error) {
	return 0, nil
}

func (m *MockICAControllerKeeper) SetTestStateRegisterInterchainAccount(returnErrForRegisterInterchainAccount bool) {
	m.returnErrForRegisterInterchainAccount = returnErrForRegisterInterchainAccount
}

func (m *MockICAControllerKeeper) SetTestStateGetInterchainAccountAddress(interchainAccountAddress string) {
	m.interchainAccountAddress = interchainAccountAddress
}
