package app

import (
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/archway-network/archway/app/upgrades"
	upgrade_0_6 "github.com/archway-network/archway/app/upgrades/06"
	upgrade10_0_0 "github.com/archway-network/archway/app/upgrades/10_0_0"
	upgrade1_0_0_rc_4 "github.com/archway-network/archway/app/upgrades/1_0_0_rc_4"
	upgrade2_0_0 "github.com/archway-network/archway/app/upgrades/2_0_0"
	upgrade3_0_0 "github.com/archway-network/archway/app/upgrades/3_0_0"
	upgrade4_0_0 "github.com/archway-network/archway/app/upgrades/4_0_0"
	upgrade4_0_2 "github.com/archway-network/archway/app/upgrades/4_0_2"
	upgrade6_0_0 "github.com/archway-network/archway/app/upgrades/6_0_0"
	upgrade7_0_0 "github.com/archway-network/archway/app/upgrades/7_0_0"
	upgrade9_0_0 "github.com/archway-network/archway/app/upgrades/9_0_0"
	upgradelatest "github.com/archway-network/archway/app/upgrades/latest"
	"github.com/archway-network/archway/app/upgrades/constantineupgrades"
)

// UPGRADES

var Upgrades = []upgrades.Upgrade{
	upgrade_0_6.Upgrade,       // v0.6.0
	upgrade1_0_0_rc_4.Upgrade, // v1.0.0-rc.4
	upgrade2_0_0.Upgrade,      // v2.0.0
	upgrade3_0_0.Upgrade,      // v3.0.0
	upgrade4_0_0.Upgrade,      // v4.0.0
	upgrade4_0_2.Upgrade,      // v4.0.2
	upgrade6_0_0.Upgrade,      // v6.0.0
	upgrade7_0_0.Upgrade,      // v7.0.0
	// upgrade8_0_0.Upgrade,  // v8.0.0: was reserved for a consensus breaking wasmd upgrade
	upgrade9_0_0.Upgrade,  // v9.0.0
	upgrade10_0_0.Upgrade, // v10.0.0
	upgradelatest.Upgrade,

	// constantine only
	constantineupgrades.WASMD_50_Amino_Patch,
}

func (app *ArchwayApp) RegisterUpgradeHandlers() {
	app.setUpgradeHandlers()
	app.setUpgradeStoreLoaders()
}

func (app *ArchwayApp) setUpgradeStoreLoaders() {
	upgradeInfo, err := app.Keepers.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}

	if app.Keepers.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
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
		app.Keepers.UpgradeKeeper.SetUpgradeHandler(
			u.UpgradeName,
			u.CreateUpgradeHandler(app.ModuleManager, app.configurator, app.Keepers),
		)
	}
}
