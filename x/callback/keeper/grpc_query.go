package keeper

import (
	"context"

	"github.com/archway-network/archway/x/callback/types"
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
func (*QueryServer) Callbacks(context.Context, *types.QueryCallbacksRequest) (*types.QueryCallbacksResponse, error) {
	panic("unimplemented 👻")
}

// EstimateCallbackFees implements types.QueryServer.
func (*QueryServer) EstimateCallbackFees(context.Context, *types.QueryEstimateCallbackFeesRequest) (*types.QueryEstimateCallbackFeesResponse, error) {
	panic("unimplemented 👻")
}

// Params implements types.QueryServer.
func (*QueryServer) Params(context.Context, *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	panic("unimplemented 👻")
}
