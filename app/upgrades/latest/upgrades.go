package upgradelatest

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/archway-network/archway/app/upgrades"
)

// This upgrade handler is used for all the current changes to the protocol

const Name = "latest"

const NameAsciiArt = `                          
             ###     ###     ### 
     # #     #       # #     # #    
     # #     ###     # #     # #   
      #        #     # #     # # 
             ###  #  ###  #  ### 

`

var Upgrade = upgrades.Upgrade{
	UpgradeName: Name,
	CreateUpgradeHandler: func(mm *module.Manager, cfg module.Configurator, accountKeeper keeper.AccountKeeper) upgradetypes.UpgradeHandler {
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
			crisistypes.ModuleName,
			consensustypes.ModuleName,
			group.ModuleName,
		},
	},
}
