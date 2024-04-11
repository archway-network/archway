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

			ctx.Logger().Info("Setting default params for the new modules")
			// Setting callback params
			callbackParams, err := keepers.CallbackKeeper.GetParams(ctx)
			if err != nil {
				return nil, err
			}
			callbackParams.CallbackGasLimit = 150000
			callbackParams.MaxBlockReservationLimit = 10
			callbackParams.MaxFutureReservationLimit = 432000 // roughly 30 days
			callbackParams.BlockReservationFeeMultiplier = sdk.MustNewDecFromStr("0.0")
			callbackParams.FutureReservationFeeMultiplier = sdk.MustNewDecFromStr("1000000000000.0")
			err = keepers.CallbackKeeper.SetParams(ctx, callbackParams)
			if err != nil {
				return nil, err
			}

			// Setting cwerrors params
			cwerrorsParams, err := keepers.CWErrorsKeeper.GetParams(ctx)
			if err != nil {
				return nil, err
			}
			cwerrorsParams.ErrorStoredTime = 302400                                                      // roughly 21 days
			cwerrorsParams.SubscriptionFee = sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000000000000000) // 1 ARCH (1e18 attoarch)
			cwerrorsParams.SubscriptionPeriod = 302400                                                   // roughly 21 days
			err = keepers.CWErrorsKeeper.SetParams(ctx, cwerrorsParams)
			if err != nil {
				return nil, err
			}

			// Setting cwica params
			cwicaParams, err := keepers.CWICAKeeper.GetParams(ctx)
			if err != nil {
				return nil, err
			}
			cwicaParams.MsgSendTxMaxMessages = 5
			err = keepers.CWICAKeeper.SetParams(ctx, cwicaParams)
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
