---
name: Release Checklist
about: Create a checklist for the upcoming release
title: 'vX.X.X: Release Checklist'
labels: release
assignees: ''

---
## Archwayd Release Checklist

This checklist is to be used for tracking the final things to do to wrap up a new release of the Archway protocol binary as well as all the post upgrade maintenance work.

### Before Release
- [ ] Ensure all the Github workflows are passing on `main`
- [ ] Rename the `latest` upgrade handlers to the `vX.X.X`. This includes:
   - [ ] Rename the package name in `app/upgrades/latest/upgrades.go` from `upgradelatest` to `upgradesX_X_X`
   - [ ] Set the name of the upgrade handler by setting the `Name` to the new version name `vX.X.X`
   - [ ] Set the `NameAsciiArt` to match what the version is expected to be. You can generate the art from [here](https://ascii-generator.site/t/) and using the font `3x5`
   - [ ] In `app/app_upgrades` change the import path from `upgradelatest "github.com/archway-network/archway/app/upgrades/latest"` to `upgradeX_X_X "github.com/archway-network/archway/app/upgrades/X_X_X"`
   - [ ]  In `app/app_upgrades` change the new Upgrade Handler reference from `upgradelatest.Upgrade` to `upgradeX_X_X.Upgrade`
- [ ] Update the `upgradeName` value to `vX.X.X` in the `interchaintest/setup.go`. This is used in the Chain Upgrade test and modifying this ensures that we simulate the upgrade accurately.  
- [ ] Update the CHANEGLOG. This includes
   - [ ] Renaming the `## [Unreleased]` header to `[vX.X.X](https://github.com/archway-network/archway/releases/tag/vX.X.X)`
   - [ ] Removed any unused headers
   - [ ] Fix any typos or duplicates
- [ ] Create a PR with the above changes titled `chore: vX release changes`

### Release
- [ ] Once above PR is merged, create a new release [here](https://github.com/archway-network/archway/releases/new). Name the release and the tag as `vX.X.X` and the contents are copy pasted from CHANGELOG.md
- [ ] At the end add a version compare link as so `**Full Changelog**: https://github.com/archway-network/archway/compare/v(X-1).00...vX.0.0`
- [ ] Ensure all release artifacts are successfully built

### Post Release
- [ ] Update CHANGELOG by adding the following to the top of the file
```markdown
## [Unreleased]

### Added

### Changed

### Deprecated

### Removed

### Fixed

### Improvements
```
- [ ] Add a placeholder upgrade handler as follows
  - [ ] Add new file at path `upgrades/latest/upgrades.go` with the following contents
```go
package upgradelatest

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

        "github.com/archway-network/archway/app/keepers"
	"github.com/archway-network/archway/app/upgrades"
)

// This upgrade handler is used for all the current changes to the protocol

const Name = "latest"
const NameAsciiArt = ""

var Upgrade = upgrades.Upgrade{
	UpgradeName: Name,
	CreateUpgradeHandler: func(mm *module.Manager, cfg module.Configurator, keepers keepers.ArchwayKeepers) upgradetypes.UpgradeHandler {
		return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			migrations, err := mm.RunMigrations(ctx, cfg, fromVM)
			if err != nil {
				return nil, err
			}

			ctx.Logger().Info(upgrades.ArchwayLogo + NameAsciiArt)
			return migrations, nil
		}
	},
	StoreUpgrades: storetypes.StoreUpgrades{},
} 
```
  - [ ] Add the latest upgrade handler reference to `app/app_upgrades.go` as the following package import `upgradelatest "github.com/archway-network/archway/app/upgrades/latest"` and the upgrade name as `upgradelatest.Upgrade, // latest - This upgrade handler is used for all the current changes to the protocol`
- [ ] Update the `initialVersion` value in the `interchaintest/setup.go` to "vX.X.X" and `upgradeName` to `latest`. This is used in the Chain Upgrade test and modifying this ensures that we always have an upgrade handler test for each PR
- [ ] Create a PR for the above with the title `chore: Post vX.X.X release maintenance`

