package upgrade06

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v4/modules/apps/29-fee/types"

	"github.com/archway-network/archway/app/keepers"
	"github.com/archway-network/archway/app/upgrades"
)

const Name = "v0.6.0"

var Upgrade = upgrades.Upgrade{
	UpgradeName: Name,
	CreateUpgradeHandler: func(mm *module.Manager, cfg module.Configurator, _ keepers.ArchwayKeepers) upgradetypes.UpgradeHandler {
		return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			return mm.RunMigrations(ctx, cfg, fromVM)
		}
	},
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{ibcfeetypes.ModuleName},
	},
}
