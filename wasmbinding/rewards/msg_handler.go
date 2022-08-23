package rewards

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/archway-network/archway/wasmbinding/rewards/types"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

// KeeperWriterExpected defines the x/rewards keeper expected write operations.
type KeeperWriterExpected interface {
	SetContractMetadata(ctx sdk.Context, senderAddr, contractAddr sdk.AccAddress, metaUpdates rewardsTypes.ContractMetadata) error
	WithdrawRewardsByRecordsLimit(ctx sdk.Context, rewardsAddr sdk.AccAddress, recordsLimit uint64) (sdk.Coins, int, error)
	WithdrawRewardsByRecordIDs(ctx sdk.Context, rewardsAddr sdk.AccAddress, recordIDs []uint64) (sdk.Coins, int, error)
}

// MsgHandler provides a custom WASM message handler for the x/rewards module.
type MsgHandler struct {
	rewardsKeeper KeeperWriterExpected
}

// NewRewardsMsgHandler creates a new MsgHandler instance.
func NewRewardsMsgHandler(rk KeeperWriterExpected) MsgHandler {
	return MsgHandler{
		rewardsKeeper: rk,
	}
}

// DispatchMsg validates and executes a custom WASM msg.
func (h MsgHandler) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg types.Msg) ([]sdk.Event, [][]byte, error) {
	// Validate the input
	if err := msg.Validate(); err != nil {
		return nil, nil, sdkErrors.Wrap(rewardsTypes.ErrInvalidRequest, fmt.Sprintf("x/rewards: sub-msg validation: %v", err))
	}

	// Execute operation (one of)
	switch {
	case msg.UpdateMetadata != nil:
		return h.updateContractMetadata(ctx, contractAddr, *msg.UpdateMetadata)
	case msg.WithdrawRewards != nil:
		return h.withdrawContractRewards(ctx, contractAddr, *msg.WithdrawRewards)
	default:
		return nil, nil, sdkErrors.Wrap(rewardsTypes.ErrInvalidRequest, "x/rewards: unknown operation")
	}
}

// updateContractMetadata updates the contract metadata.
func (h MsgHandler) updateContractMetadata(ctx sdk.Context, contractAddr sdk.AccAddress, req types.UpdateMetadataRequest) ([]sdk.Event, [][]byte, error) {
	if err := h.rewardsKeeper.SetContractMetadata(ctx, contractAddr, contractAddr, req.ToSDK()); err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}

// withdrawContractRewards withdraws the rewards for the contract address.
func (h MsgHandler) withdrawContractRewards(ctx sdk.Context, contractAddr sdk.AccAddress, req types.WithdrawRewardsRequest) ([]sdk.Event, [][]byte, error) {
	var totalRewards sdk.Coins
	var recordsUsed int
	var err error

	if req.RecordsLimit > 0 {
		totalRewards, recordsUsed, err = h.rewardsKeeper.WithdrawRewardsByRecordsLimit(ctx, contractAddr, req.RecordsLimit)
	}
	if len(req.RecordIDs) > 0 {
		totalRewards, recordsUsed, err = h.rewardsKeeper.WithdrawRewardsByRecordIDs(ctx, contractAddr, req.RecordIDs)
	}
	if err != nil {
		return nil, nil, err
	}

	resBz, err := json.Marshal(types.NewWithdrawRewardsResponse(totalRewards, recordsUsed))
	if err != nil {
		return nil, nil, fmt.Errorf("result JSON marshal: %w", err)
	}

	return nil, [][]byte{resBz}, nil
}
