package gastracker

import (
	"context"
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ gstTypes.MsgServer = &msgServer{}

func NewMsgServer(keeper GasTrackingKeeper) gstTypes.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

type msgServer struct {
	keeper GasTrackingKeeper
}

func (m *msgServer) SetContractMetadata(goCtx context.Context, msg *gstTypes.MsgSetContractMetadata) (*gstTypes.MsgSetContractMetadataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	senderAddress, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "sender")
	}

	contractAddress, err := sdk.AccAddressFromBech32(msg.ContractAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "contractAddress")
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, gstTypes.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	))

	// TODO: If not happening at earlier stage, we should check whether or not metadata is nil

	if err := m.keeper.SetContractMetadata(ctx, senderAddress, contractAddress, *msg.Metadata); err != nil {
		return nil, err
	}

	return &gstTypes.MsgSetContractMetadataResponse{}, nil

}
