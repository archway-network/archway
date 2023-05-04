package app_test

import (
	"testing"
	"time"

	storeTypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/require"

	"github.com/archway-network/archway/app/upgrades"
	e2eTesting "github.com/archway-network/archway/e2e/testing"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

func TestUpgrades(t *testing.T) {
	// create test chain
	chain := e2eTesting.NewTestChain(t, 1,
		e2eTesting.WithBlockGasLimit(100_000_000),
		e2eTesting.WithMaxWithdrawRecords(rewardsTypes.MaxWithdrawRecordsParamLimit),
	)

	// create software upgrade proposal and make it pass
	upgradeProposal := &upgradetypes.SoftwareUpgradeProposal{
		Title:       "a test upgrade",
		Description: "we're doing a test upgrade wohoo",
		Plan: upgradetypes.Plan{
			Name:   "test-upgrade",
			Height: 500,
			Info:   "some info we do not care about, right now",
		},
	}

	chain.ExecuteGovProposal(chain.GetAccount(0), true, upgradeProposal)

	chain.GoToHeight(upgradeProposal.Plan.Height-2, 1*time.Second)
	// now if we go to next height we will have a panic because of the upgrade
	require.Panics(t, func() {
		chain.NextBlock(1 * time.Second)
	})
	// add faux upgrade handler
	proposalExecuted := false
	fauxUpgrade := upgrades.Upgrade{
		UpgradeName: "test-upgrade",
		CreateUpgradeHandler: func(manager *module.Manager, configurator module.Configurator) upgradetypes.UpgradeHandler {
			return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
				proposalExecuted = true
				return manager.RunMigrations(ctx, configurator, fromVM)
			}
		},
		StoreUpgrades: storeTypes.StoreUpgrades{},
	}
	chain.GetApp().AddUpgradeHandler(fauxUpgrade)
	chain.NextBlock(1 * time.Second)
	require.True(t, proposalExecuted, "proposal was not executed")
}
