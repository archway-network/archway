package keeper

import (
	"context"

	"github.com/archway-network/archway/x/cwregistry/types"
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
func (m msgServer) RegisterCode(context.Context, *types.MsgRegisterCode) (*types.MsgRegisterCodeResponse, error) {
	panic("unimplemented")
}

// RegisterContract implements types.MsgServer.
func (m msgServer) RegisterContract(context.Context, *types.MsgRegisterContract) (*types.MsgRegisterContractResponse, error) {
	panic("unimplemented")
}
