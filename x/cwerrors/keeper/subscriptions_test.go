package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/cwerrors/types"
)

func (s *KeeperTestSuite) TestSetSubscription() {
	// Setting up chain and contract in mock wasm keeper
	ctx, keeper := s.chain.GetContext(), s.chain.GetApp().Keepers.CWErrorsKeeper
	contractViewer := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	contractAddresses := e2eTesting.GenContractAddresses(3)
	contractAddr := contractAddresses[0]
	contractAddr2 := contractAddresses[1]
	contractAdminAcc := s.chain.GetAccount(0)
	contractNotAdminAcc := s.chain.GetAccount(1)
	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)
	fees := sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)

	// TEST CASE 1: Contract does not exist
	_, err := keeper.SetSubscription(ctx, contractAdminAcc.Address, contractAddr2, fees)
	s.Require().ErrorIs(err, types.ErrContractNotFound)

	// TEST CASE 2: Sender unauthorized to set subscription
	_, err = keeper.SetSubscription(ctx, contractNotAdminAcc.Address, contractAddr, fees)
	s.Require().ErrorIs(err, types.ErrUnauthorized)

	// TEST CASE 3: Subscription fee is less than the minimum subscription fee
	params, err := keeper.GetParams(ctx)
	s.Require().NoError(err)
	err = keeper.SetParams(ctx, types.Params{
		ErrorStoredTime:    params.ErrorStoredTime,
		SubscriptionFee:    sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
		SubscriptionPeriod: params.SubscriptionPeriod,
	})
	s.Require().NoError(err)
	_, err = keeper.SetSubscription(ctx, contractAdminAcc.Address, contractAddr, fees)
	s.Require().ErrorIs(err, types.ErrInsufficientSubscriptionFee)
	err = keeper.SetParams(ctx, types.DefaultParams())
	s.Require().NoError(err)

	// TEST CASE 4: Successful subscription
	subscriptionEndHeight, err := keeper.SetSubscription(ctx, contractAdminAcc.Address, contractAddr, fees)
	s.Require().NoError(err)
	expectedEndDate := ctx.BlockHeight() + types.DefaultParams().SubscriptionPeriod
	s.Require().Equal(subscriptionEndHeight, expectedEndDate)

	// TEST CASE 5: Subscription already exists - subscription end height gets updated
	// Go to next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	subscriptionEndHeight, err = keeper.SetSubscription(ctx, contractAdminAcc.Address, contractAddr, fees)
	s.Require().NoError(err)
	s.Require().Equal(subscriptionEndHeight, expectedEndDate+1)

	// TEST CASE 6: Subscription being updated by the contract itself
	// Go to next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	subscriptionEndHeight, err = keeper.SetSubscription(ctx, contractAddr, contractAddr, fees)
	s.Require().NoError(err)
	s.Require().Equal(subscriptionEndHeight, expectedEndDate+2)
}

func (s *KeeperTestSuite) TestHasSubscription() {
	ctx, keeper := s.chain.GetContext(), s.chain.GetApp().Keepers.CWErrorsKeeper
	contractViewer := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	contractAddresses := e2eTesting.GenContractAddresses(3)
	contractAddr := contractAddresses[0]
	contractAdminAcc := s.chain.GetAccount(0)
	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)
	fees := sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)

	// TEST CASE 1: Subscription does not exist
	hasSub := keeper.HasSubscription(ctx, contractAddr)
	s.Require().False(hasSub)

	// TEST CASE 2: Subscription exists
	_, err := keeper.SetSubscription(ctx, contractAdminAcc.Address, contractAddr, fees)
	s.Require().NoError(err)
	hasSub = keeper.HasSubscription(ctx, contractAddr)
	s.Require().True(hasSub)
}

func (s *KeeperTestSuite) TestGetSubscription() {
	ctx, keeper := s.chain.GetContext(), s.chain.GetApp().Keepers.CWErrorsKeeper
	contractViewer := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	contractAddresses := e2eTesting.GenContractAddresses(3)
	contractAddr := contractAddresses[0]
	contractAdminAcc := s.chain.GetAccount(0)
	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)
	fees := sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)

	// TEST CASE 1: Subscription does not exist
	_, endHeight := keeper.GetSubscription(ctx, contractAddr)
	s.Require().Equal(endHeight, int64(0))

	// TEST CASE 2: Subscription exists
	endHeight, err := keeper.SetSubscription(ctx, contractAdminAcc.Address, contractAddr, fees)
	s.Require().NoError(err)
	found, foundEndHeight := keeper.GetSubscription(ctx, contractAddr)
	s.Require().True(found)
	s.Require().Equal(endHeight, foundEndHeight)
}

func (s *KeeperTestSuite) TestPruneSubscriptionsEndBlock() {
	ctx, keeper := s.chain.GetContext(), s.chain.GetApp().Keepers.CWErrorsKeeper
	contractViewer := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	contractAddresses := e2eTesting.GenContractAddresses(3)
	contractAddr := contractAddresses[0]
	contractAddr2 := contractAddresses[1]
	contractAddr3 := contractAddresses[2]
	contractAdminAcc := s.chain.GetAccount(0)
	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)
	contractViewer.AddContractAdmin(
		contractAddr2.String(),
		contractAdminAcc.Address.String(),
	)
	contractViewer.AddContractAdmin(
		contractAddr3.String(),
		contractAdminAcc.Address.String(),
	)

	fees := sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)

	// TEST CASE 1: No subscriptions to prune
	err := keeper.PruneSubscriptionsEndBlock(ctx)
	s.Require().NoError(err)

	// TEST CASE 2: Prune subscriptions
	endHeight, err := keeper.SetSubscription(ctx, contractAdminAcc.Address, contractAddr, fees)
	s.Require().NoError(err)
	ctx = ctx.WithBlockHeight(endHeight)
	err = keeper.PruneSubscriptionsEndBlock(ctx)
	s.Require().NoError(err)
	hasSub := keeper.HasSubscription(ctx, contractAddr)
	s.Require().False(hasSub)

	// TEST CASE 3: Prune subscriptions with multiple subscriptions
	endHeight, err = keeper.SetSubscription(ctx, contractAdminAcc.Address, contractAddr, fees)
	s.Require().NoError(err)
	_, err = keeper.SetSubscription(ctx, contractAdminAcc.Address, contractAddr2, fees)
	s.Require().NoError(err)
	_, err = keeper.SetSubscription(ctx, contractAdminAcc.Address, contractAddr3, fees)
	s.Require().NoError(err)

	//increment block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	// enxtend the subscription for contractAddr3
	_, err = keeper.SetSubscription(ctx, contractAdminAcc.Address, contractAddr3, fees)
	s.Require().NoError(err)

	ctx = ctx.WithBlockHeight(endHeight)
	err = keeper.PruneSubscriptionsEndBlock(ctx)
	s.Require().NoError(err)
	hasSub = keeper.HasSubscription(ctx, contractAddr)
	s.Require().False(hasSub)
	hasSub = keeper.HasSubscription(ctx, contractAddr2)
	s.Require().False(hasSub)
	hasSub = keeper.HasSubscription(ctx, contractAddr3)
	s.Require().True(hasSub)

	//increment block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	err = keeper.PruneSubscriptionsEndBlock(ctx)
	s.Require().NoError(err)
	hasSub = keeper.HasSubscription(ctx, contractAddr3)
	s.Require().False(hasSub)
}
