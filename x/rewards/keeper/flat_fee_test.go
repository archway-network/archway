package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

func (s *KeeperTestSuite) TestSetFlatFee() {
	ctx, keeper := s.chain.GetContext(), s.chain.GetApp().RewardsKeeper
	contractAdminAcc := s.chain.GetAccount(0)
	contractViewer := testutils.NewMockContractViewer()
	keeper.SetContractInfoViewer(contractViewer)

	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	fee := sdk.NewInt64Coin("test", 10)

	s.Run("Fail: non-existing contract metadata", func() {
		err := keeper.SetFlatFee(ctx, contractAdminAcc.Address, rewardsTypes.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         fee,
		})
		s.Assert().ErrorIs(err, rewardsTypes.ErrMetadataNotFound)
	})

	contractViewer.AddContractAdmin(contractAddr.String(), contractAdminAcc.Address.String())
	var metaCurrent rewardsTypes.ContractMetadata
	metaCurrent.ContractAddress = contractAddr.String()
	metaCurrent.OwnerAddress = contractAdminAcc.Address.String()
	_ = keeper.SetContractMetadata(ctx, contractAdminAcc.Address, contractAddr, metaCurrent)

	s.Run("OK: set flat fee", func() {
		err := keeper.SetFlatFee(ctx, contractAdminAcc.Address, rewardsTypes.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         fee,
		})
		s.Require().NoError(err)

		flatFee, ok := keeper.GetFlatFee(ctx, contractAddr)
		s.Require().True(ok)
		s.Require().Equal(fee, flatFee)
	})

	s.Run("OK: remove flat fee", func() {
		err := keeper.SetFlatFee(ctx, contractAdminAcc.Address, rewardsTypes.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         sdk.NewInt64Coin("test", 0),
		})
		s.Require().NoError(err)

		flatFee, ok := keeper.GetFlatFee(ctx, contractAddr)
		s.Require().False(ok)
		s.Require().Equal(sdk.Coin{}, flatFee)
	})
}
