package testutils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = (*MockMsg)(nil)

// MockMsg is a dummy sdk.Msg.
type MockMsg struct{}

// NewMockMsg creates a new MockMsg.
func NewMockMsg() *MockMsg {
	return &MockMsg{}
}

// Reset implements the proto.Message interface.
func (msg MockMsg) Reset() {}

// String implements the proto.Message interface.
func (msg MockMsg) String() string {
	return ""
}

// ProtoMessage implements the proto.Message interface.
func (msg MockMsg) ProtoMessage() {}

// ValidateBasic implements the sdk.Msg interface.
func (msg MockMsg) ValidateBasic() error {
	return nil
}

// GetSigners implements the sdk.Msg interface.
func (msg MockMsg) GetSigners() []sdk.AccAddress {
	return nil
}
