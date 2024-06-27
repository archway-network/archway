package keeper

import (
	"context"

	"github.com/archway-network/archway/x/cwregistry/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
func (q *QueryServer) CodeMetadata(c context.Context, req *types.QueryCodeMetadataRequest) (*types.QueryCodeMetadataResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	codeMetadata, err := q.keeper.GetCodeMetadata(ctx, req.GetCodeId())
	return &types.QueryCodeMetadataResponse{CodeMetadata: &codeMetadata}, err
}

// CodeSchema implements types.QueryServer.
func (q *QueryServer) CodeSchema(c context.Context, req *types.QueryCodeSchemaRequest) (*types.QueryCodeSchemaResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	if !q.keeper.HasCodeMetadata(ctx, req.GetCodeId()) {
		return nil, status.Error(codes.NotFound, "code metadata not found")
	}
	schema, err := q.keeper.GetSchema(req.GetCodeId())
	return &types.QueryCodeSchemaResponse{Schema: schema}, err
}

// ContractMetadata implements types.QueryServer.
func (q *QueryServer) ContractMetadata(c context.Context, req *types.QueryContractMetadataRequest) (*types.QueryContractMetadataResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	contractAddr, err := sdk.AccAddressFromBech32(req.GetContractAddress())
	if err != nil {
		return nil, err
	}
	codeInfo := q.keeper.wasmKeeper.GetContractInfo(ctx, contractAddr)
	codeMetadata, err := q.keeper.GetCodeMetadata(ctx, codeInfo.CodeID)
	return &types.QueryContractMetadataResponse{CodeMetadata: &codeMetadata}, err
}

// ContractSchema implements types.QueryServer.
func (q *QueryServer) ContractSchema(c context.Context, req *types.QueryContractSchemaRequest) (*types.QueryContractSchemaResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	contractAddr, err := sdk.AccAddressFromBech32(req.GetContractAddress())
	if err != nil {
		return nil, err
	}
	codeInfo := q.keeper.wasmKeeper.GetContractInfo(ctx, contractAddr)
	if !q.keeper.HasCodeMetadata(ctx, codeInfo.CodeID) {
		return nil, status.Error(codes.NotFound, "code metadata not found")
	}
	schema, err := q.keeper.GetSchema(codeInfo.CodeID)
	return &types.QueryContractSchemaResponse{Schema: schema}, err
}
