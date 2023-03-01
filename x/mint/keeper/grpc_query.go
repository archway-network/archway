package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/archway-network/archway/x/mint/types"
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

// Params implements the types.QueryServer interface.
func (s *QueryServer) Params(c context.Context, request *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryParamsResponse{
		Params: s.keeper.GetParams(ctx),
	}, nil
}

// Inflation implements the types.QueryServer interface.
func (s *QueryServer) Inflation(c context.Context, request *types.QueryInflationRequest) (*types.QueryInflationResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	blockInfo, found := s.keeper.GetLastBlockInfo(ctx)
	if !found {
		return nil, status.Errorf(codes.NotFound, "last block info: not found")
	}

	return &types.QueryInflationResponse{
		Inflation: blockInfo,
	}, nil
}
