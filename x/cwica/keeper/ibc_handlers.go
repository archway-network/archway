package keeper

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/archway-network/archway/x/cwica/types"
)

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

	var metadata icatypes.Metadata
	if err := icatypes.ModuleCdc.UnmarshalJSON([]byte(counterpartyVersion), &metadata); err != nil {
		return errors.Wrapf(icatypes.ErrUnknownDataType, "cannot unmarshal ICS-27 interchain accounts metadata")
	}

	successMsg := types.SudoPayload{
		Ica: &types.ICASuccess{
			AccountRegistered: &types.AccountRegistered{
				CounterpartyAddress: metadata.Address,
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

	if ack.GetError() == "" { // if no error from the counterparty chain
		sudoMsg := types.SudoPayload{
			Ica: &types.ICASuccess{
				TxExecuted: &types.TxExecuted{
					Packet: &packet,
					Data:   ack.GetResult(),
				},
			},
		}
		sudoMsgPayload, err := json.Marshal(sudoMsg)
		if err != nil {
			return fmt.Errorf("failed to marshal MessageSuccess: %v", err)
		}
		_, err = k.sudoKeeper.Sudo(ctx, contractAddress, sudoMsgPayload)
		if err != nil {
			k.Logger(ctx).Debug("HandleAcknowledgement: failed to Sudo contract on packet acknowledgement", "error", err)
		}
	} else { // if error from the counterparty chain
		packetMsg, err := json.Marshal(packet)
		if err != nil {
			return fmt.Errorf("failed to marshal packet: %v", err)
		}
		sudoerr := types.NewSudoError(types.ModuleErrors_ERR_EXEC_FAILURE, contractAddress.String(), string(packetMsg), ack.GetError())
		err = k.errorsKeeper.SetError(ctx, sudoerr)
		if err != nil {
			return fmt.Errorf("failed to set error: %v", err)
		}
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

	sudoerr := types.NewSudoError(types.ModuleErrors_ERR_PACKET_TIMEOUT, contractAddress.String(), string(packetMsg), "IBC packet timeout")
	err = k.errorsKeeper.SetError(ctx, sudoerr)
	if err != nil {
		return fmt.Errorf("failed to set error: %v", err)
	}

	return nil
}
