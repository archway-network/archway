package rewards

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	rewardsMsgTypes "github.com/archway-network/archway/wasmbinding/rewards/types"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

// KeeperWriterExpected defines the x/rewards keeper expected write operations.
type KeeperWriterExpected interface {
	SetContractMetadata(ctx sdk.Context, senderAddr, contractAddr sdk.AccAddress, metaUpdates rewardsTypes.ContractMetadata) error
	WithdrawRewardsByRecordsLimit(ctx sdk.Context, rewardsAddr sdk.AccAddress, recordsLimit uint64) (sdk.Coins, int, error)
	WithdrawRewardsByRecordIDs(ctx sdk.Context, rewardsAddr sdk.AccAddress, recordIDs []uint64) (sdk.Coins, int, error)
	SetFlatFee(ctx sdk.Context, senderAddr sdk.AccAddress, flatFeeUpdate rewardsTypes.FlatFee) error
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

// UpdateContractMetadata updates the contract metadata.
func (h MsgHandler) UpdateContractMetadata(ctx sdk.Context, senderAddr sdk.AccAddress, req rewardsMsgTypes.UpdateContractMetadataRequest) ([]sdk.Event, [][]byte, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("updateContractMetadata: %w", err)
	}

	var contractAddr sdk.AccAddress
	var isSet bool

	if contractAddr, isSet = req.MustGetContractAddressOk(); !isSet {
		contractAddr = senderAddr
	}

	if err := h.rewardsKeeper.SetContractMetadata(ctx, senderAddr, contractAddr, req.ToSDK()); err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}

// WithdrawContractRewards withdraws the rewards for the contract address.
func (h MsgHandler) WithdrawContractRewards(ctx sdk.Context, contractAddr sdk.AccAddress, req rewardsMsgTypes.WithdrawRewardsRequest) ([]sdk.Event, [][]byte, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("withdrawRewards: %w", err)
	}

	var totalRewards sdk.Coins
	var recordsUsed int
	var err error

	if req.RecordsLimit != nil {
		totalRewards, recordsUsed, err = h.rewardsKeeper.WithdrawRewardsByRecordsLimit(ctx, contractAddr, *req.RecordsLimit)
	}
	if len(req.RecordIDs) > 0 {
		totalRewards, recordsUsed, err = h.rewardsKeeper.WithdrawRewardsByRecordIDs(ctx, contractAddr, req.RecordIDs)
	}
	if err != nil {
		return nil, nil, err
	}

	resBz, err := json.Marshal(rewardsMsgTypes.NewWithdrawRewardsResponse(totalRewards, recordsUsed))
	if err != nil {
		return nil, nil, fmt.Errorf("result JSON marshal: %w", err)
	}

	return nil, [][]byte{resBz}, nil
}

// SetFlatFee sets the flat fee for the contract address.
func (h MsgHandler) SetFlatFee(ctx sdk.Context, senderAddr sdk.AccAddress, req rewardsMsgTypes.SetFlatFeeRequest) ([]sdk.Event, [][]byte, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("setFlatFee: %w", err)
	}

	if err := h.rewardsKeeper.SetFlatFee(ctx, senderAddr, req.ToSDK()); err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}
