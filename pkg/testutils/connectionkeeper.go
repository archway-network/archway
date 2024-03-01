package testutils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
)

// MockConnectionlKeeper is a mock of ConnectionKeeper interface.
type MockConnectionlKeeper struct {
	connectionFound bool
	connection      connectiontypes.ConnectionEnd
}

func NewMockConnectionlKeeper() *MockConnectionlKeeper {
	return &MockConnectionlKeeper{}
}

func (m *MockConnectionlKeeper) GetConnection(ctx sdk.Context, connectionID string) (connectiontypes.ConnectionEnd, bool) {
	return m.connection, m.connectionFound
}

func (m *MockConnectionlKeeper) SetTestStateConnection() {
	m.connection = connectiontypes.ConnectionEnd{
		Counterparty: connectiontypes.Counterparty{
			ConnectionId: "test",
		},
	}
	m.connectionFound = true
}
