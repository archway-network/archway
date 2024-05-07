package testutils

import (
	types0 "github.com/cosmos/cosmos-sdk/types"
	types4 "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	exported "github.com/cosmos/ibc-go/v8/modules/core/exported"
)

// MockChannelKeeper is a mock of ChannelKeeper interface.
type MockChannelKeeper struct {
	nextSequence uint64
}

// NewMockChannelKeeper creates a new mock instance.
func NewMockChannelKeeper() *MockChannelKeeper {
	return &MockChannelKeeper{
		nextSequence: 0,
	}
}

// GetChannel mocks base method.
func (m *MockChannelKeeper) GetChannel(ctx types0.Context, srcPort, srcChan string) (types4.Channel, bool) {
	return types4.Channel{}, true
}

// GetConnection mocks base method.
func (m *MockChannelKeeper) GetConnection(ctx types0.Context, connectionID string) (exported.ConnectionI, error) {
	return nil, nil
}

// GetNextSequenceSend mocks base method.
func (m *MockChannelKeeper) GetNextSequenceSend(ctx types0.Context, portID, channelID string) (uint64, bool) {
	if m.nextSequence != 0 {
		return m.nextSequence, true
	}
	return 0, false
}

func (m *MockChannelKeeper) SetTestStateNextSequenceSend(nextSequence uint64) {
	m.nextSequence = nextSequence
}
