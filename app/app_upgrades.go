package app

import (
	"fmt"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/archway-network/archway/app/upgrades"
	upgrade_0_6 "github.com/archway-network/archway/app/upgrades/06"
	upgrade1_0_0_rc_4 "github.com/archway-network/archway/app/upgrades/1_0_0_rc_4"
	upgrade2_0_0 "github.com/archway-network/archway/app/upgrades/2_0_0"
)

// UPGRADES

var Upgrades = []upgrades.Upgrade{
	upgrade_0_6.Upgrade,       // v0.6.0
	upgrade1_0_0_rc_4.Upgrade, // v1.0.0-rc.4
	upgrade2_0_0.Upgrade,      // v2.0.0
}

func (app *ArchwayApp) setupUpgrades() {
	app.setUpgradeHandlers()
	app.setUpgradeStoreLoaders()
}

func (app *ArchwayApp) setUpgradeStoreLoaders() {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	for _, u := range Upgrades {
		if upgradeInfo.Name == u.UpgradeName {
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &u.StoreUpgrades))
		}
	}
}

func (app *ArchwayApp) setUpgradeHandlers() {
	for _, u := range Upgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			u.UpgradeName,
			u.CreateUpgradeHandler(app.mm, app.configurator),
		)
	}
}
