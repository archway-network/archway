package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

// TestFlatFeeImportExport check flat fees import/export.
// Test updates the initial state with new records and checks that they were merged.
func (s *KeeperTestSuite) TestFlatFeeImportExport() {
	ctx, keeper := s.chain.GetContext(), s.chain.GetApp().RewardsKeeper
	contractAddrs := e2eTesting.GenContractAddresses(2)

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
		keeper.GetState().FlatFee(ctx).Import(newFlatFees)
		exportedFlatFees := keeper.GetState().FlatFee(ctx).Export()
		s.Require().NotNil(exportedFlatFees)
		s.Assert().ElementsMatch(newFlatFees, exportedFlatFees)
	})
}
