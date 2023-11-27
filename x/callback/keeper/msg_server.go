package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/callback/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
func (s MsgServer) CancelCallback(context.Context, *types.MsgCancelCallback) (*types.MsgCancelCallbackResponse, error) {
	panic("unimplemented 👻")
}

// RequestCallback implements types.MsgServer.
func (s MsgServer) RequestCallback(c context.Context, request *types.MsgRequestCallback) (*types.MsgRequestCallbackResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	sender, err := sdk.AccAddressFromBech32(request.Sender)
	if err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(c)

	zeroFee := sdk.NewInt64Coin(sdk.DefaultBondDenom, 0) // todo: fee stuff in a diff PR
	txFees := []*sdk.Coin{&zeroFee}
	blockReservationFees := []*sdk.Coin{&zeroFee}
	futureReservationFees := []*sdk.Coin{&zeroFee}
	surplusFees := []*sdk.Coin{&zeroFee}

	callback := types.NewCallback(
		request.Sender,
		request.ContractAddress,
		request.CallbackHeight,
		request.GetJobId(),
		txFees,
		blockReservationFees,
		futureReservationFees,
		surplusFees,
	)

	err = s.keeper.SaveCallback(ctx, callback)
	if err != nil {
		return &types.MsgRequestCallbackResponse{}, err
	}

	err = s.keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, request.GetFees())
	return &types.MsgRequestCallbackResponse{}, err
}

// UpdateParams implements types.MsgServer.
func (s MsgServer) UpdateParams(c context.Context, request *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	_, err := sdk.AccAddressFromBech32(request.Authority)
	if err != nil {
		return nil, err // returning error "as is" since this should not happen due to the earlier ValidateBasic call
	}

	if request.GetAuthority() != s.keeper.GetAuthority() {
		return nil, errorsmod.Wrap(types.ErrUnauthorized, "sender address is not authorized address to update module params")
	}

	err = request.GetParams().Validate() // need to explicitly validate as x/gov invokes this msg and it does not validate
	if err != nil {
		return nil, err
	}

	err = s.keeper.Params.Set(ctx, request.GetParams())
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
