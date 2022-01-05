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

	adminAddr, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "admin")
	}

	contractAddress, err := sdk.AccAddressFromBech32(msg.ContractAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "contractAddress")
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, gstTypes.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Admin),
	))

	// TODO: If not happening at earlier stage, we should check whether or not metadata is nil

	if err := m.keeper.AddContractMetadata(ctx, adminAddr, contractAddress, *msg.Metadata); err != nil {
		return nil, err
	}

	return &gstTypes.MsgSetContractMetadataResponse{}, nil

}
