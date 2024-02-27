package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/archway-network/archway/x/cwica/types"
)

var _ types.QueryServer = Keeper{}

// InterchainAccountAddress implements the Query/InterchainAccountAddress gRPC method
func (k Keeper) InterchainAccountAddress(c context.Context, req *types.QueryInterchainAccountAddressRequest) (*types.QueryInterchainAccountAddressResponse, error) {
	if req == nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	_, err := sdk.AccAddressFromBech32(req.OwnerAddress)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse address: %s", req.OwnerAddress)
	}

	portID, err := icatypes.NewControllerPortID(req.OwnerAddress)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "failed to get controller portID: %s", err)
	}

	addr, found := k.icaControllerKeeper.GetInterchainAccountAddress(ctx, req.ConnectionId, portID)
	if !found {
		return nil, errors.Wrapf(types.ErrInterchainAccountNotFound, "no interchain account found for portID %s", portID)
	}

	return &types.QueryInterchainAccountAddressResponse{InterchainAccountAddress: addr}, nil
}

// Params implements the Query/Params gRPC method
func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}
