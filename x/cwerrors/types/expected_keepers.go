package types

import (
	context "context"

	wasmdtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	rewardstypes "github.com/archway-network/archway/x/rewards/types"
)

// WasmKeeperExpected is a subset of the expected wasm keeper
type WasmKeeperExpected interface {
	// HasContractInfo returns true if the contract exists
	HasContractInfo(ctx context.Context, contractAddress sdk.AccAddress) bool
	// Sudo executes a contract message as a sudoer
	Sudo(ctx context.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
	// GetContractInfo returns the contract info
	GetContractInfo(ctx context.Context, contractAddress sdk.AccAddress) *wasmdtypes.ContractInfo
}

// BankKeeperExpected is a subset of the expected bank keeper
type BankKeeperExpected interface {
	// SendCoinsFromAccountToModule sends coins from an account to a module
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

// RewardsKeeperExpected is a subset of the expected rewards keeper
type RewardsKeeperExpected interface {
	// GetContractMetadata returns the contract metadata
	GetContractMetadata(ctx sdk.Context, contractAddr sdk.AccAddress) *rewardstypes.ContractMetadata
}
