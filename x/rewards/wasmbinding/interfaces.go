package wasmbinding

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

// ContractMetadataReader defines the GasTrackerKeeper expected operations.
type ContractMetadataReader interface {
	GetContractMetadata(ctx sdk.Context, contractAddr sdk.AccAddress) *rewardsTypes.ContractMetadata
}

// ContractMetadataWriter defines the GasTrackerKeeper expected operations.
type ContractMetadataWriter interface {
	SetContractMetadata(ctx sdk.Context, senderAddr, contractAddr sdk.AccAddress, metaUpdates rewardsTypes.ContractMetadata) error
}

type ContractMetadataReaderWriter interface {
	ContractMetadataReader
	ContractMetadataWriter
}
