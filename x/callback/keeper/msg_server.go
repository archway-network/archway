package keeper

import (
	"context"

	"github.com/archway-network/archway/x/callback/types"
)

var _ types.MsgServer = (*MsgServer)(nil)

// MsgServer implements the module gRPC messaging service.
type MsgServer struct {
	keeper Keeper
}

// NewMsgServer creates a new gRPC messaging server.
func NewMsgServer(keeper Keeper) *MsgServer {
	return &MsgServer{
		keeper: keeper,
	}
}

// CancelCallback implements types.MsgServer.
func (*MsgServer) CancelCallback(context.Context, *types.MsgCancelCallback) (*types.MsgCancelCallbackResponse, error) {
	panic("unimplemented 👻")
}

// RequestCallback implements types.MsgServer.
func (*MsgServer) RequestCallback(context.Context, *types.MsgRequestCallback) (*types.MsgRequestCallbackResponse, error) {
	panic("unimplemented 👻")
}

// UpdateParams implements types.MsgServer.
func (*MsgServer) UpdateParams(context.Context, *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	panic("unimplemented 👻")
}
