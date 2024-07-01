package upgradelatest

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/archway-network/archway/app/keepers"
	"github.com/archway-network/archway/app/upgrades"
	cwregistrytypes "github.com/archway-network/archway/x/cwregistry/types"
)

// This upgrade handler is used for all the current changes to the protocol

const Name = "latest"
const NameAsciiArt = ""

var Upgrade = upgrades.Upgrade{
	UpgradeName: Name,
	CreateUpgradeHandler: func(mm *module.Manager, cfg module.Configurator, keepers keepers.ArchwayKeepers) upgradetypes.UpgradeHandler {
		return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			migrations, err := mm.RunMigrations(ctx, cfg, fromVM)
			if err != nil {
				return nil, err
			}

			sdk.UnwrapSDKContext(ctx).Logger().Info(upgrades.ArchwayLogo + NameAsciiArt)
			return migrations, nil
		}
	},
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			cwregistrytypes.ModuleName,
		},
	},
}
