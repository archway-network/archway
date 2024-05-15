package app

import "github.com/archway-network/archway/app/upgrades"

// AddUpgradeHandler is used only for testing, and compiles as a function only in testing.
// We cannot add it to app_upgrades_test.go to avoid import cycles.
func (app *ArchwayApp) AddUpgradeHandler(upgrade upgrades.Upgrade) {
	app.Keepers.UpgradeKeeper.SetUpgradeHandler(
		upgrade.UpgradeName,
		upgrade.CreateUpgradeHandler(app.ModuleManager, app.configurator, app.Keepers),
	)
}
