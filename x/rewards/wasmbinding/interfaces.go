package wasmbinding

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

// RewardsReader defines the x/rewards keeper expected read operations.
type RewardsReader interface {
	GetContractMetadata(ctx sdk.Context, contractAddr sdk.AccAddress) *rewardsTypes.ContractMetadata
	GetRewardsRecords(ctx sdk.Context, rewardsAddr sdk.AccAddress, pageReq *query.PageRequest) ([]rewardsTypes.RewardsRecord, *query.PageResponse, error)
	MaxWithdrawRecords(ctx sdk.Context) uint64
}

// RewardsWriter defines the x/rewards keeper expected write operations.
type RewardsWriter interface {
	SetContractMetadata(ctx sdk.Context, senderAddr, contractAddr sdk.AccAddress, metaUpdates rewardsTypes.ContractMetadata) error
	WithdrawRewardsByRecordsLimit(ctx sdk.Context, rewardsAddr sdk.AccAddress, recordsLimit uint64) (sdk.Coins, int, error)
	WithdrawRewardsByRecordIDs(ctx sdk.Context, rewardsAddr sdk.AccAddress, recordIDs []uint64) (sdk.Coins, int, error)
}

type RewardsReaderWriter interface {
	RewardsReader
	RewardsWriter
}
