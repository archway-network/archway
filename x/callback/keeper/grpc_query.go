package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
func (qs *QueryServer) Callbacks(c context.Context, request *types.QueryCallbacksRequest) (*types.QueryCallbacksResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	callbacks, err := qs.keeper.GetCallbacksByHeight(ctx, request.GetBlockHeight())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not fetch the callbacks at height %d: %s", request.GetBlockHeight(), err.Error())
	}

	return &types.QueryCallbacksResponse{
		Callbacks: callbacks,
	}, nil
}

// EstimateCallbackFees implements types.QueryServer.
func (qs *QueryServer) EstimateCallbackFees(c context.Context, request *types.QueryEstimateCallbackFeesRequest) (*types.QueryEstimateCallbackFeesResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	if request.BlockHeight < ctx.BlockHeight() {
		return nil, status.Errorf(codes.InvalidArgument, "block height %d is in the past", request.BlockHeight)
	}

	params, err := qs.keeper.GetParams(ctx)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "could not fetch the module params: %s", err.Error())
	}

	futureReservationThreshold := ctx.BlockHeight() + int64(params.MaxFutureReservationLimit)
	if request.BlockHeight > futureReservationThreshold {
		return nil, status.Errorf(codes.OutOfRange, "block height %d is too far in the future. max block height can be registered %d", request.BlockHeight, futureReservationThreshold)
	}

	futureReservationFeesAmount := params.FutureReservationFeeMultiplier.MulInt64((request.GetBlockHeight() - ctx.BlockHeight()))
	futureReservationFee := sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, futureReservationFeesAmount)

	callbacksForHeight, err := qs.keeper.GetCallbacksByHeight(ctx, request.GetBlockHeight())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "could not fetch callbacks for given height: %s", err.Error())
	}
	totalCallbacks := len(callbacksForHeight)
	if totalCallbacks >= int(params.MaxBlockReservationLimit) {
		return nil, status.Errorf(codes.OutOfRange, "block height %d has reached max reservation limit", request.BlockHeight)
	}

	blockReservationFeesAmount := params.BlockReservationFeeMultiplier.MulInt64(int64(totalCallbacks))
	blockReservationFee := sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, blockReservationFeesAmount)

	transactionFeeAmount := qs.keeper.rewardsKeeper.ComputationalPriceOfGas(ctx).Amount.MulInt64(int64(params.GetCallbackGasLimit()))
	transactionFee := sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, transactionFeeAmount)

	totalFees := transactionFee.Add(blockReservationFee).Add(futureReservationFee)

	return &types.QueryEstimateCallbackFeesResponse{
		FeeSplit: &types.CallbackFeesFeeSplit{
			TransactionFees:       []*sdk.DecCoin{&transactionFee},
			BlockReservationFees:  []*sdk.DecCoin{&blockReservationFee},
			FutureReservationFees: []*sdk.DecCoin{&futureReservationFee},
		},
		TotalFees: []*sdk.DecCoin{&totalFees},
	}, nil
}

// Params implements types.QueryServer.
func (qs *QueryServer) Params(c context.Context, request *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	params, err := qs.keeper.GetParams(ctx)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "could not fetch the module params: %s", err.Error())
	}

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}
