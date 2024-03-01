package keeper

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/archway-network/archway/x/cwica/types"
)

// HandleAcknowledgement passes the acknowledgement data to the appropriate contract via a sudo call.
func (k *Keeper) HandleAcknowledgement(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	icaOwner := types.ICAOwnerFromPort(packet.SourcePort)
	contractAddress, err := sdk.AccAddressFromBech32(icaOwner)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse contract address: %s", icaOwner)
	}

	var ack channeltypes.Acknowledgement
	if err := channeltypes.SubModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return errors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-27 packet acknowledgement: %v", err)
	}

	var sudoMsgPayload []byte
	if ack.GetError() == "" { // if no error from the counterparty chain
		sudoMsg := types.SudoPayload{
			ICA: &types.MessageICASuccess{
				TxExecuted: &types.ICATxResponse{
					Data:   ack.GetResult(),
					Packet: packet,
				},
			},
		}
		sudoMsgPayload, err = json.Marshal(sudoMsg)
		if err != nil {
			return fmt.Errorf("failed to marshal MessageSuccess: %v", err)
		}
	} else { // if error from the counterparty chain
		packetMsg, err := json.Marshal(packet)
		if err != nil {
			return fmt.Errorf("failed to marshal packet: %v", err)
		}
		sudoMsg := types.SudoPayload{
			Error: types.NewSudoErrorMsg(types.SudoError{
				ErrorCode:    types.ModuleErrors_ERR_EXEC_FAILURE,
				InputPayload: string(packetMsg),
				ErrorMsg:     ack.GetError(),
			}),
		}
		sudoMsgPayload, err = json.Marshal(sudoMsg)
		if err != nil {
			return fmt.Errorf("failed to marshal MessageFailure: %v", err)
		}
	}

	_, err = k.sudoKeeper.Sudo(ctx, contractAddress, sudoMsgPayload)
	if err != nil {
		k.Logger(ctx).Debug("HandleAcknowledgement: failed to Sudo contract on packet acknowledgement", "error", err)
	}

	return nil
}

// HandleTimeout passes the timeout data to the appropriate contract via a sudo call.
// Since all ICA channels are ORDERED, a single timeout shuts down a channel.
// The channel can be reopened by registering the ICA account again.
func (k *Keeper) HandleTimeout(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	icaOwner := types.ICAOwnerFromPort(packet.SourcePort)
	contractAddress, err := sdk.AccAddressFromBech32(icaOwner)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse contract address: %s", icaOwner)
	}
	packetMsg, err := json.Marshal(packet)
	if err != nil {
		return fmt.Errorf("failed to marshal packet: %v", err)
	}
	sudoMsg := types.SudoPayload{
		Error: types.NewSudoErrorMsg(types.SudoError{
			ErrorCode:    types.ModuleErrors_ERR_PACKET_TIMEOUT,
			InputPayload: string(packetMsg),
		}),
	}
	sudoMsgPayload, err := json.Marshal(sudoMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal MessageSudoCallback: %v", err)
	}
	_, err = k.sudoKeeper.Sudo(ctx, contractAddress, sudoMsgPayload)
	if err != nil {
		k.Logger(ctx).Debug("HandleTimeout: failed to Sudo contract on packet timeout", "error", err)
	}

	return nil
}

// HandleChanOpenAck passes the data about a successfully created channel to the appropriate contract via sudo call
func (k *Keeper) HandleChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID,
	counterpartyChannelID,
	counterpartyVersion string,
) error {
	icaOwner := types.ICAOwnerFromPort(portID)
	contractAddress, err := sdk.AccAddressFromBech32(icaOwner)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse contract address: %s", icaOwner)
	}

	successMsg := types.SudoPayload{
		ICA: &types.MessageICASuccess{
			AccountRegistered: &types.OpenAckDetails{
				PortID:                portID,
				ChannelID:             channelID,
				CounterpartyChannelID: counterpartyChannelID,
				CounterpartyVersion:   counterpartyVersion,
			},
		},
	}
	sudoPayload, err := json.Marshal(successMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal MessageSuccess: %v", err)
	}

	_, err = k.sudoKeeper.Sudo(ctx, contractAddress, sudoPayload)
	if err != nil {
		k.Logger(ctx).Debug("HandleChanOpenAck: failed to sudo contract on channel open acknowledgement", "error", err)
	}

	return nil
}

// HandleChanCloseConfirm passes the data about a successfully closed channel to the appropriate contract
func (k *Keeper) HandleChanCloseConfirm(ctx sdk.Context, portID string, channelID string) error {
	icaOwner := types.ICAOwnerFromPort(portID)
	contractAddress, err := sdk.AccAddressFromBech32(icaOwner)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse contract address: %s", icaOwner)
	}

	msg := types.SudoPayload{
		ICA: &types.MessageICASuccess{
			AccountClosed: &types.ChannelClosed{
				PortID:    portID,
				ChannelID: channelID,
			},
		},
	}
	sudoPayload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal MessageSuccess: %v", err)
	}

	_, err = k.sudoKeeper.Sudo(ctx, contractAddress, sudoPayload)
	if err != nil {
		k.Logger(ctx).Debug("HandleChanCloseConfirm: failed to sudo contract on channel close confirmation", "error", err)
	}
	return nil
}
