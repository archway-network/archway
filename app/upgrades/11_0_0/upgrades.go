package upgrade11_0_0

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	usdcmigrationtypes "github.com/disperze/usdc-migration/v8/x/usdcmigration/types"

	"github.com/archway-network/archway/app/keepers"
	"github.com/archway-network/archway/app/upgrades"
)

const Name = "v11.0.0"
const NameAsciiArt = `
              #   #    ###     ###
     # #     ##  ##    # #     # #
     # #      #   #    # #     # #
      #       #   #    # #     # #
             ### ###  #  ###  #  ###

`

// NobleChannel is the Archway end of the transfer channel to Noble (noble-1
// channel-12). It is a var so local testnets can override it with
// -ldflags "-X github.com/archway-network/archway/app/upgrades/11_0_0.NobleChannel=channel-0".
var NobleChannel = "channel-29"

var Upgrade = upgrades.Upgrade{
	UpgradeName: Name,
	CreateUpgradeHandler: func(mm *module.Manager, cfg module.Configurator, keepers keepers.ArchwayKeepers) upgradetypes.UpgradeHandler {
		return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			migrations, err := mm.RunMigrations(ctx, cfg, fromVM)
			if err != nil {
				return nil, err
			}

			// RunMigrations runs InitGenesis of x/usdcmigration with its default (inert)
			// params, so the Noble channel is set afterwards. From this height on, Noble
			// USDC can no longer be received from Noble or sent out of Archway.
			params := usdcmigrationtypes.DefaultParams()
			params.NobleChannels = []string{NobleChannel}
			if err := keepers.UsdcMigrationKeeper.SetParams(ctx, params); err != nil {
				return nil, err
			}

			sdkCtx := sdk.UnwrapSDKContext(ctx)
			sdkCtx.Logger().Info("usdcmigration enabled", "noble_channels", params.NobleChannels)
			sdkCtx.Logger().Info(upgrades.ArchwayLogo + NameAsciiArt)
			return migrations, nil
		}
	},
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{usdcmigrationtypes.StoreKey},
	},
}
