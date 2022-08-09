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
func CustomMessageDecorator(gtKeeper ContractMetadataWriter) func(old wasmKeeper.Messenger) wasmKeeper.Messenger {
	return func(old wasmKeeper.Messenger) wasmKeeper.Messenger {
		return NewMsgPlugin(old, gtKeeper)
	}
}

// MsgPlugin provides custom WASM message handlers.
type MsgPlugin struct {
	rewardsKeeper    ContractMetadataWriter
	wrappedMessenger wasmKeeper.Messenger
}

// NewMsgPlugin creates a new MsgPlugin.
func NewMsgPlugin(wrappedMessenger wasmKeeper.Messenger, rk ContractMetadataWriter) *MsgPlugin {
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

		// Execute custom msgs one by one
		var resEvents []sdk.Event
		var resData [][]byte
		if customMsg.UpdateMetadata != nil {
			if err := p.updateContractMetadata(ctx, contractAddr, *customMsg.UpdateMetadata); err != nil {
				return nil, nil, fmt.Errorf("updateMetadata: %w", err)
			}
		}

		return resEvents, resData, nil
	}

	return p.wrappedMessenger.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
}

// updateContractMetadata updates the contract metadata.
func (p MsgPlugin) updateContractMetadata(ctx sdk.Context, contractAddr sdk.AccAddress, req types.UpdateMetadataRequest) error {
	return p.rewardsKeeper.SetContractMetadata(ctx, contractAddr, contractAddr, req.ToMetadata())
}
