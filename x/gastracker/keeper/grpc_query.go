package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gstTypes "github.com/archway-network/archway/x/gastracker"
)

var _ gstTypes.QueryServer = &queryServer{}

type queryServer struct {
	keeper Keeper
}

func NewQueryServer(keeper Keeper) gstTypes.QueryServer {
	return &queryServer{
		keeper: keeper,
	}
}

func (g *queryServer) ContractMetadata(ctx context.Context, request *gstTypes.QueryContractMetadataRequest) (*gstTypes.QueryContractMetadataResponse, error) {
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

func (g *queryServer) BlockGasTracking(ctx context.Context, request *gstTypes.QueryBlockGasTrackingRequest) (*gstTypes.QueryBlockGasTrackingResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	blockGasTracking := g.keeper.GetCurrentBlockTracking(sdk.UnwrapSDKContext(ctx))

	return &gstTypes.QueryBlockGasTrackingResponse{
		BlockGasTracking: &blockGasTracking,
	}, nil
}
