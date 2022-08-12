package wasmbinding

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

// RewardsReader defines the x/rewards keeper expected read operations.
type RewardsReader interface {
	GetContractMetadata(ctx sdk.Context, contractAddr sdk.AccAddress) *rewardsTypes.ContractMetadata
	GetCurrentRewards(ctx sdk.Context, rewardsAddr sdk.AccAddress) sdk.Coins
}

// RewardsWriter defines the x/rewards keeper expected write operations.
type RewardsWriter interface {
	SetContractMetadata(ctx sdk.Context, senderAddr, contractAddr sdk.AccAddress, metaUpdates rewardsTypes.ContractMetadata) error
	WithdrawRewards(ctx sdk.Context, rewardsAddr sdk.AccAddress) sdk.Coins
}

type RewardsReaderWriter interface {
	RewardsReader
	RewardsWriter
}
