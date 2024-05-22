package upgrade4_0_2_test

import (
	"fmt"
	"testing"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/suite"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
)

type UpgradeTestSuite struct {
	suite.Suite

	archway *e2eTesting.TestChain
}

func (s *UpgradeTestSuite) SetupTest() {
	s.archway = e2eTesting.NewTestChain(s.T(), 1)
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

const (
	dummyUpgradeHeight = 5
)

func (suite *UpgradeTestSuite) TestUpgrade() {
	testCases := []struct {
		name         string
		pre_upgrade  func()
		post_upgrade func()
	}{
		{
			"Feecollector does not have burn permissions, we ensure upgrade happens and account gets the burn permissions",
			func() {
				accountKeeper := suite.archway.GetApp().Keepers.AccountKeeper
				fcAccount := accountKeeper.GetModuleAccount(suite.archway.GetContext(), authtypes.FeeCollectorName)

				account, ok := fcAccount.(*authtypes.ModuleAccount)
				suite.Require().True(ok)
				account.Permissions = []string{}
				accountKeeper.SetModuleAccount(suite.archway.GetContext(), account)

				fcAccount = accountKeeper.GetModuleAccount(suite.archway.GetContext(), authtypes.FeeCollectorName)
				suite.Require().False(fcAccount.HasPermission(authtypes.Burner))
			},
			func() {
				accountKeeper := suite.archway.GetApp().Keepers.AccountKeeper
				fcAccount := accountKeeper.GetModuleAccount(suite.archway.GetContext(), authtypes.FeeCollectorName)
				suite.Require().True(fcAccount.HasPermission(authtypes.Burner))
			},
		},
		{
			"Feecollector already has burn permissions, we ensure upgrade happens smoothly",
			func() {
				accountKeeper := suite.archway.GetApp().Keepers.AccountKeeper
				fcAccount := accountKeeper.GetModuleAccount(suite.archway.GetContext(), authtypes.FeeCollectorName)
				suite.Require().True(fcAccount.HasPermission(authtypes.Burner))
			},
			func() {
				accountKeeper := suite.archway.GetApp().Keepers.AccountKeeper
				fcAccount := accountKeeper.GetModuleAccount(suite.archway.GetContext(), authtypes.FeeCollectorName)
				suite.Require().True(fcAccount.HasPermission(authtypes.Burner))
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.pre_upgrade()

			ctx := suite.archway.GetContext().WithBlockHeight(dummyUpgradeHeight - 1)
			plan := upgradetypes.Plan{Name: "v4.0.2", Height: dummyUpgradeHeight}
			upgradekeeper := suite.archway.GetApp().Keepers.UpgradeKeeper
			err := upgradekeeper.ScheduleUpgrade(ctx, plan)
			suite.Require().NoError(err)
			_, err = upgradekeeper.GetUpgradePlan(ctx)
			suite.Require().NoError(err)
			ctx = ctx.WithBlockHeight(dummyUpgradeHeight)
			suite.Require().NotPanics(func() {
				_, err = suite.archway.GetApp().BeginBlocker(ctx)
				suite.Require().NoError(err)
			})

			tc.post_upgrade()
		})
	}
}
