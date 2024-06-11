package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

func (s *KeeperTestSuite) TestSetFlatFee() {
	contractAdminAcc := testutils.AccAddress()

	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	fee := sdk.NewInt64Coin("test", 10)

	s.Run("Fail: non-existing contract metadata", func() {
		err := s.keeper.SetFlatFee(s.ctx, contractAdminAcc, rewardsTypes.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         fee,
		})
		s.Assert().ErrorIs(err, rewardsTypes.ErrMetadataNotFound)
	})

	s.wasmKeeper.AddContractAdmin(contractAddr.String(), contractAdminAcc.String())
	var metaCurrent rewardsTypes.ContractMetadata
	metaCurrent.ContractAddress = contractAddr.String()
	metaCurrent.OwnerAddress = contractAdminAcc.String()
	_ = s.keeper.SetContractMetadata(s.ctx, contractAdminAcc, contractAddr, metaCurrent)

	s.Run("Fail: rewards address not set", func() {
		err := s.keeper.SetFlatFee(s.ctx, contractAdminAcc, rewardsTypes.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         fee,
		})
		s.Assert().ErrorIs(err, rewardsTypes.ErrMetadataNotFound)
	})

	metaCurrent.RewardsAddress = contractAdminAcc.String()
	_ = s.keeper.SetContractMetadata(s.ctx, contractAdminAcc, contractAddr, metaCurrent)

	s.Run("OK: set flat fee", func() {
		err := s.keeper.SetFlatFee(s.ctx, contractAdminAcc, rewardsTypes.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         fee,
		})
		s.Require().NoError(err)

		flatFee, ok := s.keeper.GetFlatFee(s.ctx, contractAddr)
		s.Require().True(ok)
		s.Require().Equal(fee, flatFee)
	})

	s.Run("OK: remove flat fee", func() {
		err := s.keeper.SetFlatFee(s.ctx, contractAdminAcc, rewardsTypes.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         sdk.NewInt64Coin("test", 0),
		})
		s.Require().NoError(err)

		flatFee, ok := s.keeper.GetFlatFee(s.ctx, contractAddr)
		s.Require().False(ok)
		s.Require().Equal(sdk.Coin{}, flatFee)
	})
}
