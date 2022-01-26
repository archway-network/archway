package gastracker

import (
	"context"
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ gstTypes.QueryServer = &grpcQuerier{}

type grpcQuerier struct {
	keeper GasTrackingKeeper
}

func NewGRPCQuerier(keeper GasTrackingKeeper) gstTypes.QueryServer {
	return &grpcQuerier{
		keeper: keeper,
	}
}

func (g *grpcQuerier) ContractMetadata(ctx context.Context, request *gstTypes.QueryContractMetadataRequest) (*gstTypes.QueryContractMetadataResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	contractAddr, err := sdk.AccAddressFromBech32(request.Address)
	if err != nil {
		return nil, err
	}

	resp, err := g.keeper.GetContractMetadata(sdk.UnwrapSDKContext(ctx), contractAddr)
	if err != nil {
		return nil, err
	}

	return &gstTypes.QueryContractMetadataResponse{
		Metadata: &resp,
	}, nil
}

func (g *grpcQuerier) BlockGasTracking(ctx context.Context, request *gstTypes.QueryBlockGasTrackingRequest) (*gstTypes.QueryBlockGasTrackingResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	blockGasTracking, err := g.keeper.GetCurrentBlockTracking(sdk.UnwrapSDKContext(ctx))
	if err != nil {
		return nil, err
	}

	return &gstTypes.QueryBlockGasTrackingResponse{
		BlockGasTracking: &blockGasTracking,
	}, nil
}
