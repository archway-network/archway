package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// WasmKeeperExpected is a subset of the expected wasm keeper
type WasmKeeperExpected interface {
	// HasContractInfo returns true if the contract exists
	HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool
	// Sudo executes a contract message as a sudoer
	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
}

// BankKeeperExpected is a subset of the expected bank keeper
type BankKeeperExpected interface {
	// SendCoinsFromAccountToModule sends coins from an account to a module
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}
