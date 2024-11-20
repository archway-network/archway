package upgrade8_0_0

import (
	"context"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/archway-network/archway/app/keepers"
	"github.com/archway-network/archway/app/upgrades"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

const Name = "v10.0.0"
const NameAsciiArt = `
              #      ###     ###
     # #     ##      # #     # #
     # #      #      # #     # #
      #       #      # #     # #
             ###  #  ###  #  ###

`

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
}
