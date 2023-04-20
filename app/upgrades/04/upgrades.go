package upgrade03

import (
	"github.com/archway-network/archway/app/upgrades"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const Name = "v0.4.0"

var Upgrade = upgrades.Upgrade{
	UpgradeName: Name,
	CreateUpgradeHandler: func(mm *module.Manager, cfg module.Configurator) upgradetypes.UpgradeHandler {
		return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			return mm.RunMigrations(ctx, cfg, fromVM)
		}
	},
	StoreUpgrades: storetypes.StoreUpgrades{},
}
