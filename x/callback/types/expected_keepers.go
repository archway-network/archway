package types

import (
	wasmdtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	cwerrortypes "github.com/archway-network/archway/x/cwerrors/types"
	rewardstypes "github.com/archway-network/archway/x/rewards/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type WasmKeeperExpected interface {
	HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool
	GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmdtypes.ContractInfo
	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
}

type RewardsKeeperExpected interface {
	GetContractMetadata(ctx sdk.Context, contractAddr sdk.AccAddress) *rewardstypes.ContractMetadata
	ComputationalPriceOfGas(ctx sdk.Context) sdk.DecCoin
}

type BankKeeperExpected interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
	BlockedAddr(addr sdk.AccAddress) bool
}

type ErrorsKeeperExpected interface {
	SetError(ctx sdk.Context, sudoErr cwerrortypes.SudoError) error
}
