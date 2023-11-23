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
func (*QueryServer) EstimateCallbackFees(context.Context, *types.QueryEstimateCallbackFeesRequest) (*types.QueryEstimateCallbackFeesResponse, error) {
	panic("unimplemented 👻")
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
