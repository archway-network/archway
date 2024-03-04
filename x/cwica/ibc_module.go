package cwica

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/archway-network/archway/x/cwica/keeper"
)

var _ porttypes.IBCModule = IBCModule{}

// IBCModule implements the ICS26 interface for interchain accounts controller chains
type IBCModule struct {
	keeper keeper.Keeper
}

// NewIBCModule creates a new IBCModule given the keeper
func NewIBCModule(k keeper.Keeper) IBCModule {
	return IBCModule{keeper: k}
}

// OnChanOpenInit implements the IBCModule interface. We don't need to implement this handler.
func (im IBCModule) OnChanOpenInit(_ sdk.Context, _ channeltypes.Order, _ []string, _ string, _ string, _ *capabilitytypes.Capability, _ channeltypes.Counterparty, version string) (string, error) {
	return version, nil
}

// OnChanOpenAck implements the IBCModule interface. This handler is called after we create an
// account on the counterparty chain (because icaControllerKeeper.RegisterInterchainAccount opens a channel).
func (im IBCModule) OnChanOpenAck(ctx sdk.Context, portID, channelID, counterPartyChannelID, counterpartyVersion string) error {
	return im.keeper.HandleChanOpenAck(ctx, portID, channelID, counterPartyChannelID, counterpartyVersion)
}

// OnChanCloseConfirm implements the IBCModule interface.
func (im IBCModule) OnChanCloseConfirm(ctx sdk.Context, portID string, channelID string) error {
	return im.keeper.HandleChanCloseConfirm(ctx, portID, channelID)
}

// OnAcknowledgementPacket implements the IBCModule interface.
func (im IBCModule) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	return im.keeper.HandleAcknowledgement(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCModule interface.
func (im IBCModule) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	return im.keeper.HandleTimeout(ctx, packet, relayer)
}

// OnChanOpenTry implements the IBCModule interface. We don't need to implement this handler.
func (im IBCModule) OnChanOpenTry(_ sdk.Context, _ channeltypes.Order, _ []string, _, _ string, _ *capabilitytypes.Capability, _ channeltypes.Counterparty, _ string) (string, error) {
	panic("NOT NEEDED FOR CONTROLLER MODULE")
}

// OnChanOpenConfirm implements the IBCModule interface. We don't need to implement this handler.
func (im IBCModule) OnChanOpenConfirm(_ sdk.Context, _, _ string) error {
	panic("NOT NEEDED FOR CONTROLLER MODULE")
}

// OnChanCloseInit implements the IBCModule interface. We don't need to implement this handler.
func (im IBCModule) OnChanCloseInit(_ sdk.Context, _, _ string) error {
	panic("NOT NEEDED FOR CONTROLLER MODULE")
}

// OnRecvPacket implements the IBCModule interface
func (im IBCModule) OnRecvPacket(_ sdk.Context, _ channeltypes.Packet, _ sdk.AccAddress) ibcexported.Acknowledgement {
	panic("NOT NEEDED FOR CONTROLLER MODULE")
}
