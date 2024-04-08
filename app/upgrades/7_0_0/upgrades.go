package upgrade7_0_0

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"

	"github.com/archway-network/archway/app/keepers"
	"github.com/archway-network/archway/app/upgrades"
	callbackTypes "github.com/archway-network/archway/x/callback/types"
	cwerrorstypes "github.com/archway-network/archway/x/cwerrors/types"
	"github.com/archway-network/archway/x/cwfees"
	cwicatypes "github.com/archway-network/archway/x/cwica/types"
)

// This upgrade handler is used for all the current changes to the protocol

const Name = "v7.0.0"
const NameAsciiArt = `                          
             ###     ###     ### 
     # #       #     # #     # #    
     # #       #     # #     # #   
      #        #     # #     # # 
               #  #  ###  #  ### 

`

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
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			callbackTypes.ModuleName,
			cwfees.ModuleName,
			cwerrorstypes.ModuleName,
			icacontrollertypes.StoreKey,
			cwicatypes.ModuleName,
		},
	},
}
