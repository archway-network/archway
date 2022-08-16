package wasmbinding

import (
	"encoding/json"
	"fmt"

	wasmKeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmVmTypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
	"github.com/archway-network/archway/x/rewards/wasmbinding/types"
)

var _ wasmKeeper.Messenger = &MsgPlugin{}

// CustomMessageDecorator creates a new CustomQueryPlugin for WASM bindings.
func CustomMessageDecorator(gtKeeper RewardsWriter) func(old wasmKeeper.Messenger) wasmKeeper.Messenger {
	return func(old wasmKeeper.Messenger) wasmKeeper.Messenger {
		return NewMsgPlugin(old, gtKeeper)
	}
}

// MsgPlugin provides custom WASM message handlers.
type MsgPlugin struct {
	rewardsKeeper    RewardsWriter
	wrappedMessenger wasmKeeper.Messenger
}

// NewMsgPlugin creates a new MsgPlugin.
func NewMsgPlugin(wrappedMessenger wasmKeeper.Messenger, rk RewardsWriter) *MsgPlugin {
	return &MsgPlugin{
		wrappedMessenger: wrappedMessenger,
		rewardsKeeper:    rk,
	}
}

// DispatchMsg validates and executes a custom WASM msg.
func (p MsgPlugin) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmVmTypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
	if msg.Custom != nil {
		// Parse and validate the input
		var customMsg types.Msg
		if err := json.Unmarshal(msg.Custom, &customMsg); err != nil {
			return nil, nil, sdkErrors.Wrap(rewardsTypes.ErrInvalidRequest, fmt.Sprintf("custom msg JSON unmarshal: %v", err))
		}
		if err := customMsg.Validate(); err != nil {
			return nil, nil, sdkErrors.Wrap(rewardsTypes.ErrInvalidRequest, fmt.Sprintf("custom msg validation: %v", err))
		}

		// Execute custom msg (one of)
		switch {
		case customMsg.UpdateMetadata != nil:
			return p.updateContractMetadata(ctx, contractAddr, *customMsg.UpdateMetadata)
		case customMsg.WithdrawRewards != nil:
			return p.withdrawContractRewards(ctx, contractAddr, *customMsg.WithdrawRewards)
		default:
			return nil, nil, sdkErrors.Wrap(rewardsTypes.ErrInvalidRequest, "unknown request")
		}
	}

	return p.wrappedMessenger.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
}

// updateContractMetadata updates the contract metadata.
func (p MsgPlugin) updateContractMetadata(ctx sdk.Context, contractAddr sdk.AccAddress, req types.UpdateMetadataRequest) ([]sdk.Event, [][]byte, error) {
	if err := p.rewardsKeeper.SetContractMetadata(ctx, contractAddr, contractAddr, req.ToSDK()); err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}

// withdrawContractRewards withdraws the rewards for the contract address.
func (p MsgPlugin) withdrawContractRewards(ctx sdk.Context, contractAddr sdk.AccAddress, req types.WithdrawRewardsRequest) ([]sdk.Event, [][]byte, error) {
	var totalRewards sdk.Coins
	var recordsUsed int
	var err error

	if req.RecordsLimit > 0 {
		totalRewards, recordsUsed, err = p.rewardsKeeper.WithdrawRewardsByRecordsLimit(ctx, contractAddr, req.RecordsLimit)
	}
	if len(req.RecordIDs) > 0 {
		totalRewards, recordsUsed, err = p.rewardsKeeper.WithdrawRewardsByRecordIDs(ctx, contractAddr, req.RecordIDs)
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
