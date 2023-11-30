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
		return nil, status.Error(codes.InvalidArgument, "invalid contract address: "+err.Error())
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

	blockRewards, err := s.keeper.BlockRewards.Get(ctx, uint64(height))
	if err == nil {
		// it makes sense to report a height only if it does exist
		// if err != nil, it means that the block rewards tracking for the current block does not exist
		blockRewards.Height = ctx.BlockHeight()
	}
	txRewards, err := s.keeper.GetTxRewardsByBlock(ctx, uint64(height))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "block rewards tracking for the block %d: not found", height)
	}

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
		UndistributedFunds: s.keeper.UndistributedRewardsPool(ctx),
		TreasuryFunds:      s.keeper.TreasuryPool(ctx),
	}, nil
}

// EstimateTxFees implements the types.QueryServer interface.
func (s *QueryServer) EstimateTxFees(c context.Context, request *types.QueryEstimateTxFeesRequest) (*types.QueryEstimateTxFeesResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	fees := sdk.NewCoins()
	computationalPoG := s.keeper.ComputationalPriceOfGas(ctx)
	fees = fees.Add(sdk.NewCoin(computationalPoG.Denom, computationalPoG.Amount.MulInt(sdk.NewIntFromUint64(request.GasLimit)).RoundInt()))

	if request.ContractAddress != "" { // if contract address is passed in, get the flat fee and add that.
		contractAddr, err := sdk.AccAddressFromBech32(request.ContractAddress)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid contract address: "+err.Error())
		}
		contractFlatFee, found := s.keeper.GetFlatFee(ctx, contractAddr)
		if found {
			fees = fees.Add(contractFlatFee)
		}
	}

	return &types.QueryEstimateTxFeesResponse{
		GasUnitPrice: computationalPoG,
		EstimatedFee: fees.Sort(),
	}, nil
}

// OutstandingRewards implements the types.QueryServer interface.
func (s *QueryServer) OutstandingRewards(c context.Context, request *types.QueryOutstandingRewardsRequest) (*types.QueryOutstandingRewardsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	rewardsAddr, err := sdk.AccAddressFromBech32(request.RewardsAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid rewards address: "+err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)

	totalRewards := sdk.NewCoins()
	records, err := s.keeper.GetRewardsRecordsByWithdrawAddress(ctx, rewardsAddr)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "outstanding rewards for the address %s: not found", rewardsAddr)
	}
	for _, record := range records {
		totalRewards = totalRewards.Add(record.Rewards...)
	}

	return &types.QueryOutstandingRewardsResponse{
		TotalRewards: totalRewards,
		RecordsNum:   uint64(len(records)),
	}, nil
}

// RewardsRecords implements the types.QueryServer interface.
func (s *QueryServer) RewardsRecords(c context.Context, request *types.QueryRewardsRecordsRequest) (*types.QueryRewardsRecordsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	rewardsAddr, err := sdk.AccAddressFromBech32(request.RewardsAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid rewards address: "+err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)

	records, pageResp, err := s.keeper.GetRewardsRecords(ctx, rewardsAddr, request.Pagination)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "pagination request: "+err.Error())
	}

	return &types.QueryRewardsRecordsResponse{
		Records:    records,
		Pagination: pageResp,
	}, nil
}

// FlatFee implements the types.QueryServer interface.
func (s *QueryServer) FlatFee(c context.Context, request *types.QueryFlatFeeRequest) (*types.QueryFlatFeeResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	contractAddr, err := sdk.AccAddressFromBech32(request.ContractAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid contract address: "+err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)

	fee, ok := s.keeper.GetFlatFee(ctx, contractAddr)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "flat fee: not found")
	}

	return &types.QueryFlatFeeResponse{
		FlatFeeAmount: fee,
	}, nil
}
