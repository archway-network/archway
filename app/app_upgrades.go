package app

import (
	"fmt"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/archway-network/archway/app/upgrades"
	upgrade_0_3 "github.com/archway-network/archway/app/upgrades/03"
	upgrade_0_4 "github.com/archway-network/archway/app/upgrades/04"
	upgrade_0_5_2 "github.com/archway-network/archway/app/upgrades/052"
	upgrade_0_6 "github.com/archway-network/archway/app/upgrades/06"
)

// UPGRADES

var Upgrades = []upgrades.Upgrade{
	upgrade_0_3.Upgrade,   // v0.3.0
	upgrade_0_4.Upgrade,   // v0.4.0
	upgrade_0_5_2.Upgrade, // v0.5.2
	upgrade_0_6.Upgrade,   // v0.6.0
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
