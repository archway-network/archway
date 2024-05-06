package testutils

import (
	"errors"

	types0 "github.com/cosmos/cosmos-sdk/types"
	types2 "github.com/cosmos/cosmos-sdk/x/capability/types"
	types3 "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
)

// MockICAControllerKeeper is a mock of ICAControllerKeeper interface.
type MockICAControllerKeeper struct {
	returnErrForRegisterInterchainAccount bool
	interchainAccountAddress              string
	channelId                             string
	packetSequence                        uint64
}

// NewMockICAControllerKeeper creates a new mock instance.
func NewMockICAControllerKeeper() *MockICAControllerKeeper {
	return &MockICAControllerKeeper{
		returnErrForRegisterInterchainAccount: true,
		interchainAccountAddress:              "",
		channelId:                             "",
		packetSequence:                        0,
	}
}

// GetActiveChannelID mocks base method.
func (m *MockICAControllerKeeper) GetActiveChannelID(ctx types0.Context, connectionID, portID string) (string, bool) {
	if m.channelId != "" {
		return m.channelId, true
	}
	return "", false
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
	if m.packetSequence != 0 {
		return m.packetSequence, nil
	}
	return 0, errors.New("failed to send tx")
}

func (m *MockICAControllerKeeper) SetTestStateRegisterInterchainAccount(returnErrForRegisterInterchainAccount bool) {
	m.returnErrForRegisterInterchainAccount = returnErrForRegisterInterchainAccount
}

func (m *MockICAControllerKeeper) SetTestStateGetInterchainAccountAddress(interchainAccountAddress string) {
	m.interchainAccountAddress = interchainAccountAddress
}

func (m *MockICAControllerKeeper) SetTestStateGetActiveChannelID(channelId string) {
	m.channelId = channelId
}

func (m *MockICAControllerKeeper) SetTestStateSendTx(packetSequence uint64) {
	m.packetSequence = packetSequence
}
