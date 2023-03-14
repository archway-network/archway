package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	trackingTypes "github.com/archway-network/archway/x/tracking/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// ContractInfoReader defines the interface for the x/wasmd module dependency.
type ContractInfoReader interface {
	GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmTypes.ContractInfo
}

// TrackingKeeper defines the interface for the x/tracking module dependency.
type TrackingKeeper interface {
	GetCurrentTxID(ctx sdk.Context) uint64
	GetBlockTrackingInfo(ctx sdk.Context, height int64) trackingTypes.BlockTracking
	RemoveBlockTrackingInfo(ctx sdk.Context, height int64)
}

// AuthKeeper defines the interface for the x/auth module dependency.
type AuthKeeper interface {
	GetModuleAccount(ctx sdk.Context, name string) authTypes.ModuleAccountI
}

// BankKeeper defines the interface for the x/bank module dependency.
type BankKeeper interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule string, recipientModule string, amt sdk.Coins) error
}

// MintKeeper defines the interface for the x/mint module dependency.
type MintKeeper interface {
	// GetInflationForRecipient gets the sdk.Coin distributed to the given module in the current block
	GetInflationForRecipient(ctx sdk.Context, recipientName string) (sdk.Coin, bool)
}
