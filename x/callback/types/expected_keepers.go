package types

import (
	wasmdtypes "github.com/CosmWasm/wasmd/x/wasm/types"

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
