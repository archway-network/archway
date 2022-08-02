package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/archway-network/archway/x/rewards/types"
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

	meta := s.keeper.GetContractMetadata(ctx, contractAddr)
	if meta == nil {
		return nil, status.Errorf(codes.NotFound, "metadata for the contract: not found")
	}

	return &types.QueryContractMetadataResponse{
		Metadata: *meta,
	}, nil
}

// BlockRewardsTracking implements the types.QueryServer interface.
func (s *QueryServer) BlockRewardsTracking(c context.Context, request *types.QueryBlockRewardsTrackingRequest) (*types.QueryBlockRewardsTrackingResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	height := ctx.BlockHeight()

	blockRewards, found := s.keeper.state.BlockRewardsState(ctx).GetBlockRewards(height)
	if !found {
		blockRewards.Height = ctx.BlockHeight()
	}
	txRewards := s.keeper.state.TxRewardsState(ctx).GetTxRewardsByBlock(height)

	return &types.QueryBlockRewardsTrackingResponse{
		Block: types.BlockTracking{
			InflationRewards: blockRewards,
			TxRewards:        txRewards,
		},
	}, nil
}

// RewardsPool implements the types.QueryServer interface.
func (s *QueryServer) RewardsPool(c context.Context, request *types.QueryRewardsPoolRequest) (*types.QueryRewardsPoolResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryRewardsPoolResponse{
		Funds: s.keeper.UndistributedRewardsPool(ctx),
	}, nil
}
