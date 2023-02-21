package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/archway-network/archway/x/rewards/types"
)

var _ types.MsgServer = (*MsgServer)(nil)

// MsgServer implements the module gRPC messaging service.
type MsgServer struct {
	keeper Keeper
}

// NewMsgServer creates a new gRPC messaging server.
func NewMsgServer(keeper Keeper) *MsgServer {
	return &MsgServer{
		keeper: keeper,
	}
}

// SetContractMetadata implements the types.MsgServer interface.
func (s MsgServer) SetContractMetadata(c context.Context, request *types.MsgSetContractMetadata) (*types.MsgSetContractMetadataResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	senderAddr, err := sdk.AccAddressFromBech32(request.SenderAddress)
	if err != nil {
		return nil, err // returning error "as is" since this should not happen due to the earlier ValidateBasic call
	}

	contractAddr, err := sdk.AccAddressFromBech32(request.Metadata.ContractAddress)
	if err != nil {
		return nil, err // returning error "as is" since this should not happen due to the earlier ValidateBasic call
	}

	if err := s.keeper.SetContractMetadata(ctx, senderAddr, contractAddr, request.Metadata); err != nil {
		return nil, err
	}

	return &types.MsgSetContractMetadataResponse{}, nil
}

// WithdrawRewards implements the types.MsgServer interface.
func (s MsgServer) WithdrawRewards(c context.Context, request *types.MsgWithdrawRewards) (*types.MsgWithdrawRewardsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	rewardsAddr, err := sdk.AccAddressFromBech32(request.RewardsAddress)
	if err != nil {
		return nil, err // returning error "as is" since this should not happen due to the earlier ValidateBasic call
	}

	var totalRewards sdk.Coins
	var recordsUsed int

	switch modeReq := request.Mode.(type) {
	case *types.MsgWithdrawRewards_RecordsLimit_:
		totalRewards, recordsUsed, err = s.keeper.WithdrawRewardsByRecordsLimit(ctx, rewardsAddr, modeReq.RecordsLimit.Limit)
	case *types.MsgWithdrawRewards_RecordIds:
		totalRewards, recordsUsed, err = s.keeper.WithdrawRewardsByRecordIDs(ctx, rewardsAddr, modeReq.RecordIds.Ids)
	default:
		// Should never happen since the BasicValidate function checks this case
		return nil, status.Error(codes.InvalidArgument, "invalid request mode")
	}

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &types.MsgWithdrawRewardsResponse{
		RecordsNum:   uint64(recordsUsed),
		TotalRewards: totalRewards,
	}, nil
}

// SetFlatFee implements the types.MsgServer interface.
func (s MsgServer) SetFlatFee(c context.Context, request *types.MsgSetFlatFee) (*types.MsgSetFlatFeeResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	senderAddress, err := sdk.AccAddressFromBech32(request.SenderAddress)
	if err != nil {
		return nil, err // returning error "as is" since this should not happen due to the earlier ValidateBasic call
	}

	_, err = sdk.AccAddressFromBech32(request.ContractAddress)
	if err != nil {
		return nil, err // returning error "as is" since this should not happen due to the earlier ValidateBasic call
	}

	if err := s.keeper.SetFlatFee(ctx, senderAddress, types.FlatFee{
		ContractAddress: request.GetContractAddress(),
		FlatFee:         request.GetFlatFeeAmount(),
	}); err != nil {
		return nil, err
	}

	return &types.MsgSetFlatFeeResponse{}, nil
}
