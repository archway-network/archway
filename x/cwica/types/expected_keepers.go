package types

import (
	context "context"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	cwerrortypes "github.com/archway-network/archway/x/cwerrors/types"
)

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sk.AccountI
}

// WasmKeeper defines the expected interface needed to interact with the wasm module.
type WasmKeeper interface {
	HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool
	GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo
	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
}

// ICAControllerKeeper defines the expected interface needed to interact with the interchain accounts module.
type ICAControllerKeeper interface {
	GetActiveChannelID(ctx sdk.Context, connectionID, portID string) (string, bool)
	GetInterchainAccountAddress(ctx sdk.Context, connectionID, portID string) (string, bool)
	RegisterInterchainAccount(ctx sdk.Context, connectionID, owner, version string) error
	SendTx(ctx sdk.Context, chanCap *capabilitytypes.Capability, connectionID, portID string, icaPacketData icatypes.InterchainAccountPacketData, timeoutTimestamp uint64) (uint64, error)
}

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
	GetConnection(ctx sdk.Context, connectionID string) (ibcexported.ConnectionI, error)
}

// ConnectionKeeper defines the expected IBC connection keeper
type ConnectionKeeper interface {
	GetConnection(ctx sdk.Context, connectionID string) (connectiontypes.ConnectionEnd, bool)
}

// ErrorsKeeper defines the expected interface needed to interact with the cwerrors module.
type ErrorsKeeper interface {
	// SetError records a sudo error for a contract
	SetError(ctx sdk.Context, sudoErr cwerrortypes.SudoError) error
}
