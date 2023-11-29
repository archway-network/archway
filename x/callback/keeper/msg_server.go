package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
func (s MsgServer) CancelCallback(c context.Context, request *types.MsgCancelCallback) (*types.MsgCancelCallbackResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	// If a callback with same job id does not exist, return error
	callback, err := s.keeper.GetCallback(ctx, request.GetCallbackHeight(), request.GetContractAddress(), request.GetJobId())
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrCallbackNotFound, "callback with given job id does not exist for given height")
	}

	// Returning the transaction fees as the callback was never executed
	txFee := callback.GetFeeSplit().GetTransactionFees()
	err = s.keeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(request.Sender), sdk.NewCoins(*txFee))
	if err != nil {
		return nil, err
	}

	// todo: deal with the rest of the fees. later in diff pr

	// Deleting the callback from state
	err = s.keeper.DeleteCallback(ctx, request.Sender, request.GetCallbackHeight(), request.GetContractAddress(), request.GetJobId())
	if err != nil {
		return nil, err
	}
	return &types.MsgCancelCallbackResponse{
		Refund: *txFee,
	}, nil
}

// RequestCallback implements types.MsgServer.
func (s MsgServer) RequestCallback(c context.Context, request *types.MsgRequestCallback) (*types.MsgRequestCallbackResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	// Get the expected fees which is to be paid
	futureReservationFee, blockReservationFee, transactionFee, err := s.keeper.EstimateCallbackFees(ctx, request.GetCallbackHeight())
	if err != nil {
		return nil, err
	}
	expectedFees := transactionFee.Add(blockReservationFee).Add(futureReservationFee)

	// If the fees sent by the sender is less than the expected fees, return error
	if request.GetFees().IsLT(expectedFees) {
		return nil, errorsmod.Wrapf(types.ErrInsufficientFees, "expected %s, got %s", expectedFees, request.GetFees())
	}
	surplusFees := request.GetFees().Sub(expectedFees) // Calculating any surplus user has sent

	// Send the fees into module account
	err = s.keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(request.Sender), types.ModuleName, sdk.NewCoins(request.GetFees()))
	if err != nil {
		return nil, err
	}

	// Save the callback in state
	callback := types.NewCallback(
		request.Sender,
		request.ContractAddress,
		request.CallbackHeight,
		request.GetJobId(),
		transactionFee,
		blockReservationFee,
		futureReservationFee,
		surplusFees,
	)
	err = s.keeper.SaveCallback(ctx, callback)
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
		return nil, err
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
