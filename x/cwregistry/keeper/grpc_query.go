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
