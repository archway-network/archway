package keeper

import (
	"context"

	"github.com/archway-network/archway/x/cwregistry/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type msgServer struct {
	Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// RegisterCode implements types.MsgServer.
func (m msgServer) RegisterCode(c context.Context, req *types.MsgRegisterCode) (*types.MsgRegisterCodeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, err
	}
	codeMetadata := types.CodeMetadata{
		Source:        req.SourceMetadata,
		SourceBuilder: req.SourceBuilder,
		Schema:        req.Schema,
		Contacts:      req.Contacts,
	}
	err = m.SetCodeMetadata(ctx, sender, req.CodeId, codeMetadata)
	return &types.MsgRegisterCodeResponse{}, err
}
