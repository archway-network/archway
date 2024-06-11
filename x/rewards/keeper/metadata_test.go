package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

func (s *KeeperTestSuite) TestSetContractMetadata() {
	contractAdminAcc, otherAcc := testutils.AccAddress(), testutils.AccAddress()
	rewardAddr := sdk.AccAddress{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	contractAddr := e2eTesting.GenContractAddresses(1)[0]

	s.Run("Fail: non-existing contract", func() {
		err := s.keeper.SetContractMetadata(s.ctx, otherAcc, contractAddr, rewardsTypes.ContractMetadata{})
		s.Assert().ErrorIs(err, rewardsTypes.ErrContractNotFound)
	})

	// Set contract admin
	s.wasmKeeper.AddContractAdmin(contractAddr.String(), contractAdminAcc.String())

	var metaCurrent rewardsTypes.ContractMetadata
	s.Run("OK: create", func() {
		metaCurrent.ContractAddress = contractAddr.String()
		metaCurrent.OwnerAddress = contractAdminAcc.String()
		metaCurrent.RewardsAddress = rewardAddr.String()

		err := s.keeper.SetContractMetadata(s.ctx, contractAdminAcc, contractAddr, metaCurrent)
		s.Require().NoError(err)

		metaReceived := s.keeper.GetContractMetadata(s.ctx, contractAddr)
		s.Require().NotNil(metaReceived)
		s.Assert().Equal(metaCurrent, *metaReceived)
	})

	s.Run("Fail: not a contract admin", func() {
		metaCurrent := metaCurrent
		metaCurrent.OwnerAddress = otherAcc.String()
		err := s.keeper.SetContractMetadata(s.ctx, otherAcc, contractAddr, metaCurrent)
		s.Assert().ErrorIs(err, rewardsTypes.ErrUnauthorized)
	})

	s.Run("OK: set RewardsAddr", func() {
		metaCurrent.RewardsAddress = otherAcc.String()

		err := s.keeper.SetContractMetadata(s.ctx, contractAdminAcc, contractAddr, metaCurrent)
		s.Require().NoError(err)

		metaReceived := s.keeper.GetContractMetadata(s.ctx, contractAddr)
		s.Require().NotNil(metaReceived)
		s.Assert().Equal(metaCurrent, *metaReceived)
	})

	s.Run("OK: update OwnerAddr (change ownership)", func() {
		metaCurrent.OwnerAddress = otherAcc.String()

		err := s.keeper.SetContractMetadata(s.ctx, contractAdminAcc, contractAddr, metaCurrent)
		s.Require().NoError(err)

		metaReceived := s.keeper.GetContractMetadata(s.ctx, contractAddr)
		s.Require().NotNil(metaReceived)
		s.Assert().Equal(metaCurrent, *metaReceived)
	})

	s.Run("Fail: try to regain ownership", func() {
		metaCurrent.OwnerAddress = contractAdminAcc.String()

		err := s.keeper.SetContractMetadata(s.ctx, contractAdminAcc, contractAddr, metaCurrent)
		s.Require().ErrorIs(err, rewardsTypes.ErrUnauthorized)
	})

	s.Run("Fail: unable to set reward address to a module account", func() {
		metaCurrent.RewardsAddress = authtypes.NewModuleAddress("distribution").String()
		err := s.keeper.SetContractMetadata(s.ctx, contractAdminAcc, contractAddr, metaCurrent)
		s.Require().ErrorIs(err, rewardsTypes.ErrInvalidRequest)
	})
}
