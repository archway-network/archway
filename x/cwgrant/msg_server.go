package cwgrant

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/cwgrant/types"
)

var _ types.MsgServer = (*msgServer)(nil)

func NewMsgServer(k Keeper) types.MsgServer { return msgServer{k: k} }

type msgServer struct{ k Keeper }

func (m msgServer) RegisterAsGranter(ctx context.Context, msg *types.MsgRegisterAsGranter) (*types.MsgRegisterAsGranterResponse, error) {
	granterAddr, err := sdk.AccAddressFromBech32(msg.GrantingContract)
	if err != nil {
		return nil, err
	}
	return &types.MsgRegisterAsGranterResponse{}, m.k.RegisterAsGranter(ctx, granterAddr)
}

func (m msgServer) UnregisterAsGranter(ctx context.Context, msg *types.MsgUnregisterAsGranter) (*types.MsgUnregisterAsGranterResponse, error) {
	granterAddr, err := sdk.AccAddressFromBech32(msg.GrantingContract)
	if err != nil {
		return nil, err
	}

	return &types.MsgUnregisterAsGranterResponse{}, m.k.UnregisterAsGranter(ctx, granterAddr)
}
