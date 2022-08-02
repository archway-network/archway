package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/archway-network/archway/x/tracking/types"
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

// BlockGasTracking implements the types.QueryServer interface.
func (s *QueryServer) BlockGasTracking(c context.Context, request *types.QueryBlockGasTrackingRequest) (*types.QueryBlockGasTrackingResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	blockInfo := s.keeper.GetBlockTrackingInfo(ctx, ctx.BlockHeight())

	return &types.QueryBlockGasTrackingResponse{
		Block: blockInfo,
	}, nil
}
