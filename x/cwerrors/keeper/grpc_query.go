package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/archway-network/archway/x/cwerrors/types"
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

// Errors implements types.QueryServer.
func (qs *QueryServer) Errors(c context.Context, request *types.QueryErrorsRequest) (*types.QueryErrorsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	addr, err := sdk.AccAddressFromBech32(request.ContractAddress)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid contract address: %s", err.Error())
	}

	errors, err := qs.keeper.GetErrorsByContractAddress(sdk.UnwrapSDKContext(c), addr)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "could not fetch the errors: %s", err.Error())
	}

	return &types.QueryErrorsResponse{
		Errors: errors,
	}, nil
}

// IsSubscribed implements types.QueryServer.
func (qs *QueryServer) IsSubscribed(c context.Context, request *types.QueryIsSubscribedRequest) (*types.QueryIsSubscribedResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	contractAddr, err := sdk.AccAddressFromBech32(request.ContractAddress)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid contract address: %s", err.Error())
	}
	hasSub, validtill := qs.keeper.GetSubscription(sdk.UnwrapSDKContext(c), contractAddr)
	return &types.QueryIsSubscribedResponse{Subscribed: hasSub, SubscriptionValidTill: validtill}, nil
}

// Params implements types.QueryServer.
func (qs *QueryServer) Params(c context.Context, request *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	params, err := qs.keeper.GetParams(sdk.UnwrapSDKContext(c))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "could not fetch the module params: %s", err.Error())
	}

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}
