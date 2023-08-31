package upgrade4_0_1_test

import (
	"fmt"
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

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
				accountKeeper := suite.archway.GetApp().AccountKeeper
				fcAccount := accountKeeper.GetModuleAccount(suite.archway.GetContext(), authtypes.FeeCollectorName)

				account, ok := fcAccount.(*authtypes.ModuleAccount)
				suite.Require().True(ok)
				account.Permissions = []string{}
				accountKeeper.SetModuleAccount(suite.archway.GetContext(), account)

				fcAccount = accountKeeper.GetModuleAccount(suite.archway.GetContext(), authtypes.FeeCollectorName)
				suite.Require().False(fcAccount.HasPermission(authtypes.Burner))
			},
			func() {
				accountKeeper := suite.archway.GetApp().AccountKeeper
				fcAccount := accountKeeper.GetModuleAccount(suite.archway.GetContext(), authtypes.FeeCollectorName)
				suite.Require().True(fcAccount.HasPermission(authtypes.Burner))
			},
		},
		{
			"Feecollector already has burn permissions, we ensure upgrade happens smoothly",
			func() {
				accountKeeper := suite.archway.GetApp().AccountKeeper
				fcAccount := accountKeeper.GetModuleAccount(suite.archway.GetContext(), authtypes.FeeCollectorName)
				suite.Require().True(fcAccount.HasPermission(authtypes.Burner))
			},
			func() {
				accountKeeper := suite.archway.GetApp().AccountKeeper
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
			plan := upgradetypes.Plan{Name: "v4.0.1", Height: dummyUpgradeHeight}
			upgradekeeper := suite.archway.GetApp().UpgradeKeeper
			err := upgradekeeper.ScheduleUpgrade(ctx, plan)
			suite.Require().NoError(err)
			_, exists := upgradekeeper.GetUpgradePlan(ctx)
			suite.Require().True(exists)
			ctx = ctx.WithBlockHeight(dummyUpgradeHeight)
			suite.Require().NotPanics(func() {
				suite.archway.GetApp().BeginBlocker(ctx, abci.RequestBeginBlock{})
			})

			tc.post_upgrade()
		})
	}
}
