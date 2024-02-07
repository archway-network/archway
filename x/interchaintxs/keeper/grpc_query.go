package keeper

import (
	"context"

	"cosmossdk.io/errors"
	"github.com/archway-network/archway/x/interchaintxs/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) InterchainAccountAddress(c context.Context, req *types.QueryInterchainAccountAddressRequest) (*types.QueryInterchainAccountAddressResponse, error) {
	if req == nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	icaOwner, err := types.NewICAOwner(req.OwnerAddress, req.InterchainAccountId)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "failed to create ica owner: %s", err)
	}

	portID, err := icatypes.NewControllerPortID(icaOwner.String())
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "failed to get controller portID: %s", err)
	}

	addr, found := k.icaControllerKeeper.GetInterchainAccountAddress(ctx, req.ConnectionId, portID)
	if !found {
		return nil, errors.Wrapf(types.ErrInterchainAccountNotFound, "no interchain account found for portID %s", portID)
	}

	return &types.QueryInterchainAccountAddressResponse{InterchainAccountAddress: addr}, nil
}

func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}
