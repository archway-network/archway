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
