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
		err := keeper.SetFlatFee(ctx, contractAddr, fee)
		s.Assert().ErrorIs(err, rewardsTypes.ErrMetadataNotFound)
	})

	contractViewer.AddContractAdmin(contractAddr.String(), contractAdminAcc.Address.String())
	var metaCurrent rewardsTypes.ContractMetadata
	metaCurrent.ContractAddress = contractAddr.String()
	metaCurrent.OwnerAddress = contractAdminAcc.Address.String()
	_ = keeper.SetContractMetadata(ctx, contractAdminAcc.Address, contractAddr, metaCurrent)

	s.Run("OK: set flat fee", func() {
		err := keeper.SetFlatFee(ctx, contractAddr, fee)
		s.Require().NoError(err)

		flatFee, ok := keeper.GetFlatFee(ctx, contractAddr)
		s.Require().True(ok)
		s.Require().Equal(fee, flatFee)
	})

	s.Run("OK: remove flat fee", func() {
		err := keeper.SetFlatFee(ctx, contractAddr, sdk.NewInt64Coin("test", 0))
		s.Require().NoError(err)

		flatFee, ok := keeper.GetFlatFee(ctx, contractAddr)
		s.Require().False(ok)
		s.Require().Equal(sdk.Coin{}, flatFee)
	})
}

// TestFlatFeeImportExport check flat fees import/export.
// Test updates the initial state with new records and checks that they were merged.
func (s *KeeperTestSuite) TestFlatFeeImportExportSuccess() {
	ctx, keeper := s.chain.GetContext(), s.chain.GetApp().RewardsKeeper
	contractAddrs := e2eTesting.GenContractAddresses(2)
	accAddrs, _ := e2eTesting.GenAccounts(2)
	newMetadata := []rewardsTypes.ContractMetadata{
		{
			ContractAddress: contractAddrs[0].String(),
			OwnerAddress:    accAddrs[0].String(),
		},
		{
			ContractAddress: contractAddrs[1].String(),
			OwnerAddress:    accAddrs[1].String(),
			RewardsAddress:  accAddrs[1].String(),
		},
	}
	keeper.GetState().ContractMetadataState(ctx).Import(newMetadata)

	newFlatFees := []rewardsTypes.FlatFee{
		{
			ContractAddress: contractAddrs[0].String(),
			FlatFee:         sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),
		},
		{
			ContractAddress: contractAddrs[1].String(),
			FlatFee:         sdk.NewCoin("uarch", sdk.NewInt(1)),
		},
	}

	s.Run("Check import export of flat fees", func() {
		keeper.ImportFlatFees(ctx, newFlatFees)
		genesisStateReceived := keeper.ExportGenesis(ctx)
		s.Require().NotNil(genesisStateReceived)
		s.Assert().ElementsMatch(newFlatFees, genesisStateReceived.FlatFees)
	})
}
