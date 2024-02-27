package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	types2 "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/cwica/types"
)

func (s *KeeperTestSuite) TestKeeper_InterchainAccountAddress() {
	ctx, keeper := s.chain.GetContext().WithBlockHeight(1), s.chain.GetApp().Keepers.CWICAKeeper
	wmKeeper, icaCtrlKeeper := testutils.NewMockContractViewer(), testutils.NewMockICAControllerKeeper()
	keeper.SetWasmKeeper(wmKeeper)
	keeper.SetICAControllerKeeper(icaCtrlKeeper)
	contractAdminAcc := s.chain.GetAccount(0)
	wctx := sdk.WrapSDKContext(ctx)

	resp, err := keeper.InterchainAccountAddress(wctx, nil)
	s.Require().ErrorIs(err, sdkerrors.ErrInvalidRequest)
	s.Require().Nil(resp)

	resp, err = keeper.InterchainAccountAddress(wctx, &types.QueryInterchainAccountAddressRequest{
		OwnerAddress: "nonbech32",
		ConnectionId: "connection-0",
	})
	s.Require().ErrorContains(err, "failed to parse address")
	s.Require().Nil(resp)

	portID := fmt.Sprintf("%s%s.%s", types2.ControllerPortPrefix, contractAdminAcc.Address.String())
	addr, found := icaCtrlKeeper.GetInterchainAccountAddress(ctx, "connection-0", portID)
	s.Require().False(found)
	s.Require().Equal("", addr)
	resp, err = keeper.InterchainAccountAddress(wctx, &types.QueryInterchainAccountAddressRequest{
		OwnerAddress: contractAdminAcc.Address.String(),
		ConnectionId: "connection-0",
	})
	s.Require().ErrorContains(err, "no interchain account found for portID")
	s.Require().Nil(resp)

	icaCtrlKeeper.SetTestStateGetInterchainAccountAddress(contractAdminAcc.Address.String())
	portID = fmt.Sprintf("%s%s.%s", types2.ControllerPortPrefix, contractAdminAcc.Address.String(), "test1")
	addr, found = icaCtrlKeeper.GetInterchainAccountAddress(ctx, "connection-0", portID)
	s.Require().True(found)
	s.Require().Equal(contractAdminAcc.Address.String(), addr)
	resp, err = keeper.InterchainAccountAddress(wctx, &types.QueryInterchainAccountAddressRequest{
		OwnerAddress: contractAdminAcc.Address.String(),
		ConnectionId: "connection-0",
	})
	s.Require().NoError(err)
	s.Require().Equal(&types.QueryInterchainAccountAddressResponse{InterchainAccountAddress: contractAdminAcc.Address.String()}, resp)
}

func (s *KeeperTestSuite) TestParamsQuery() {
	ctx, keeper := s.chain.GetContext().WithBlockHeight(1), s.chain.GetApp().Keepers.CWICAKeeper
	wctx := sdk.WrapSDKContext(ctx)
	params := types.DefaultParams()
	err := keeper.SetParams(ctx, params)
	s.Require().NoError(err)

	response, err := keeper.Params(wctx, nil)
	s.Require().Error(err)
	s.Require().Nil(response)

	response, err = keeper.Params(wctx, &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(&types.QueryParamsResponse{Params: params}, response)
}
