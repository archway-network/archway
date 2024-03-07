package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/archway-network/archway/x/cwerrors/types"
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

// SubscribeToError implements types.MsgServer.
func (s *MsgServer) SubscribeToError(c context.Context, request *types.MsgSubscribeToError) (*types.MsgSubscribeToErrorResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	contractAddr, err := sdk.AccAddressFromBech32(request.Contract)
	if err != nil {
		return nil, err
	}

	subscriptionEndHeight, err := s.keeper.SetSubscription(sdk.UnwrapSDKContext(c), contractAddr, request.Fee)
	if err != nil {
		return nil, err
	}
	return &types.MsgSubscribeToErrorResponse{SubscriptionValidTill: subscriptionEndHeight}, nil
}

// UpdateParams implements types.MsgServer.
func (s MsgServer) UpdateParams(c context.Context, request *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	_, err := sdk.AccAddressFromBech32(request.Authority)
	if err != nil {
		return nil, err
	}

	if request.GetAuthority() != s.keeper.GetAuthority() {
		return nil, errorsmod.Wrap(types.ErrUnauthorized, "sender address is not authorized address to update module params")
	}

	err = request.GetParams().Validate() // need to explicitly validate as x/gov invokes this msg and it does not validate
	if err != nil {
		return nil, err
	}

	err = s.keeper.Params.Set(sdk.UnwrapSDKContext(c), request.GetParams())
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
