package constantineupgrades

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/archway-network/archway/app/keepers"
	"github.com/archway-network/archway/app/upgrades"
	"github.com/cosmos/cosmos-sdk/types/module"
)

var WASMD_50_Amino_Patch = upgrades.Upgrade{
	UpgradeName: "wasmd_50_amino_patch",
	CreateUpgradeHandler: func(mm *module.Manager, cfg module.Configurator, _ keepers.ArchwayKeepers) upgradetypes.UpgradeHandler {
		return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			return mm.RunMigrations(ctx, cfg, fromVM)
		}
	},
	StoreUpgrades: storetypes.StoreUpgrades{},
}
