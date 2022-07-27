package keeper

import (
	"context"
	"github.com/archway-network/archway/x/rewards/types"
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

// ContractMetadata implements the types.QueryServer interface.
func (s *QueryServer) ContractMetadata(c context.Context, request *types.QueryContractMetadataRequest) (*types.QueryContractMetadataResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	contractAddr, err := sdk.AccAddressFromBech32(request.ContractAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid contract address")
	}

	ctx := sdk.UnwrapSDKContext(c)
	metaState := s.keeper.state.ContractMetadataState(ctx)

	meta, found := metaState.GetContractMetadata(contractAddr)
	if !found {
		return nil, status.Errorf(codes.NotFound, "metadata for the contract: not found")
	}

	return &types.QueryContractMetadataResponse{
		Metadata: meta,
	}, nil
}
