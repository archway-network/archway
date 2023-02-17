package wasmbinding

import (
	"encoding/json"
	"fmt"

	wasmKeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmVmTypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/archway-network/archway/wasmbinding/rewards"
	"github.com/archway-network/archway/wasmbinding/types"
)

var _ wasmKeeper.Messenger = MsgDispatcher{}

// MsgDispatcher dispatches custom WASM queries.
type MsgDispatcher struct {
	rewardsHandler   rewards.MsgHandler
	wrappedMessenger wasmKeeper.Messenger
}

// NewMsgDispatcher creates a new MsgDispatcher instance.
func NewMsgDispatcher(wrappedMessenger wasmKeeper.Messenger, rh rewards.MsgHandler) MsgDispatcher {
	return MsgDispatcher{
		wrappedMessenger: wrappedMessenger,
		rewardsHandler:   rh,
	}
}

// DispatchMsg validates and executes a custom WASM msg.
func (d MsgDispatcher) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmVmTypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
	// Skip non-custom message
	if msg.Custom == nil {
		return d.wrappedMessenger.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
	}

	// Parse and validate the input
	var customMsg types.Msg
	if err := json.Unmarshal(msg.Custom, &customMsg); err != nil {
		return nil, nil, sdkErrors.Wrap(sdkErrors.ErrInvalidRequest, fmt.Sprintf("custom msg JSON unmarshal: %v", err))
	}
	if err := customMsg.Validate(); err != nil {
		return nil, nil, sdkErrors.Wrap(sdkErrors.ErrInvalidRequest, fmt.Sprintf("custom msg validation: %v", err))
	}

	// Execute custom sub-msg (one of)
	switch {
	case customMsg.UpdateContractMetadata != nil:
		return d.rewardsHandler.UpdateContractMetadata(ctx, contractAddr, *customMsg.UpdateContractMetadata)
	case customMsg.WithdrawRewards != nil:
		return d.rewardsHandler.WithdrawContractRewards(ctx, contractAddr, *customMsg.WithdrawRewards)
	case customMsg.SetFlatFee != nil:
		return d.rewardsHandler.SetFlatFee(ctx, contractAddr, *customMsg.SetFlatFee)
	default:
		// That should never happen, since we validate the input above
		return nil, nil, sdkErrors.Wrap(wasmdTypes.ErrUnknownMsg, "no custom handler found")
	}
}
