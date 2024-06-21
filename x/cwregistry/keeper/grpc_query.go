package keeper

import (
	"context"

	"github.com/archway-network/archway/x/cwregistry/types"
)

var _ types.QueryServer = &QueryServer{}

type QueryServer struct {
	keeper Keeper
}

// NewQueryServer creates a new gRPC query server.
func NewQueryServer(keeper Keeper) *QueryServer {
	return &QueryServer{
		keeper: keeper,
	}
}

// CodeMetadata implements types.QueryServer.
func (q *QueryServer) CodeMetadata(context.Context, *types.QueryCodeMetadataRequest) (*types.QueryCodeMetadataResponse, error) {
	panic("unimplemented")
}

// CodeSchema implements types.QueryServer.
func (q *QueryServer) CodeSchema(context.Context, *types.QueryCodeSchemaRequest) (*types.QueryCodeSchemaResponse, error) {
	panic("unimplemented")
}

// ContractMetadata implements types.QueryServer.
func (q *QueryServer) ContractMetadata(context.Context, *types.QueryContractMetadataRequest) (*types.QueryContractMetadataResponse, error) {
	panic("unimplemented")
}

// ContractSchema implements types.QueryServer.
func (q *QueryServer) ContractSchema(context.Context, *types.QueryContractSchemaRequest) (*types.QueryContractSchemaResponse, error) {
	panic("unimplemented")
}
