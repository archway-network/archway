package keeper

import (
	"context"

	"github.com/archway-network/archway/x/callback/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = &QueryServer{}

// QueryServer implements the module gRPC query service.
type QueryServer struct {
	keeper Keeper
}

// NewQueryServer creates a new gRPC query server.
func NewQueryServer(keeper Keeper) *QueryServer {
	return &QueryServer{
		keeper: keeper,
	}
}

// Callbacks implements types.QueryServer.
func (s *QueryServer) Callbacks(c context.Context, request *types.QueryCallbacksRequest) (*types.QueryCallbacksResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	callbacks, err := s.keeper.GetCallbacksByHeight(ctx, request.GetBlockHeight())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not fetch the callbacks at height %d: %s", request.GetBlockHeight(), err.Error())
	}

	return &types.QueryCallbacksResponse{
		Callbacks: callbacks,
	}, nil
}

// EstimateCallbackFees implements types.QueryServer.
func (s *QueryServer) EstimateCallbackFees(c context.Context, request *types.QueryEstimateCallbackFeesRequest) (*types.QueryEstimateCallbackFeesResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	if request.BlockHeight < ctx.BlockHeight() {
		return nil, status.Errorf(codes.InvalidArgument, "block height %d is in the past", request.BlockHeight)
	}

	params, err := s.keeper.GetParams(ctx)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "could not fetch the module params: %s", err.Error())
	}

	reservationThresholdHeight := ctx.BlockHeight() + int64(params.MaxFutureReservationLimit)
	if request.BlockHeight > reservationThresholdHeight {
		return nil, status.Errorf(codes.OutOfRange, "block height %d is too far in the future. max block height can be registered %d", request.BlockHeight, reservationThresholdHeight)
	}

	// transactionFee := get param gas limit and multiply by estimate-fees
	// futureREservationFees := get block height diff and multiply by multiplier
	// blockReservationFees := get number of callbacks in block and how many pending and multiply by multiplier

	futureReservationFeesAmount := params.FutureReservationFeeMultiplier.MulInt64((request.GetBlockHeight() - ctx.BlockHeight()))
	futureReservationFees := sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, futureReservationFeesAmount)

	return &types.QueryEstimateCallbackFeesResponse{
		FeeSplit: &types.CallbackFeesFeeSplit{
			TransactionFees:       nil,
			BlockReservationFees:  nil,
			FutureReservationFees: futureReservationFees,
		},
		TotalFees: nil,
	}, nil
}

// Params implements types.QueryServer.
func (s *QueryServer) Params(c context.Context, request *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	params, err := s.keeper.GetParams(ctx)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "could not fetch the module params: %s", err.Error())
	}

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}
