package keeper

import (
	"context"

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
	panic("unimplemented 👻")
}
