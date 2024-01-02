package cwgrant

import (
	"context"

	"github.com/archway-network/archway/x/cwgrant/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
