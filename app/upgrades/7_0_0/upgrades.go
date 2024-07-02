package upgrade7_0_0

import (
	"context"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"

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
		return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			migrations, err := mm.RunMigrations(ctx, cfg, fromVM)
			if err != nil {
				return nil, err
			}

			sdk.UnwrapSDKContext(ctx).Logger().Info("Setting default params for the new modules")
			bondDenom, err := keepers.StakingKeeper.BondDenom(ctx)
			if err != nil {
				return nil, err
			}
			unwrappedCtx := sdk.UnwrapSDKContext(ctx)
			// Setting callback params
			callbackParams, err := keepers.CallbackKeeper.GetParams(unwrappedCtx)
			if err != nil {
				return nil, err
			}
			callbackParams.CallbackGasLimit = 150000
			callbackParams.MaxBlockReservationLimit = 10
			callbackParams.MaxFutureReservationLimit = 432000 // roughly 30 days
			callbackParams.BlockReservationFeeMultiplier = math.LegacyMustNewDecFromStr("0.0")
			callbackParams.FutureReservationFeeMultiplier = math.LegacyMustNewDecFromStr("1000000000000.0")
			err = keepers.CallbackKeeper.SetParams(unwrappedCtx, callbackParams)
			if err != nil {
				return nil, err
			}

			// Setting cwerrors params
			cwerrorsParams, err := keepers.CWErrorsKeeper.GetParams(unwrappedCtx)
			if err != nil {
				return nil, err
			}
			cwerrorsParams.ErrorStoredTime = 302400                                           // roughly 21 days
			cwerrorsParams.SubscriptionFee = sdk.NewInt64Coin(bondDenom, 1000000000000000000) // 1 ARCH (1e18 attoarch)
			cwerrorsParams.SubscriptionPeriod = 302400                                        // roughly 21 days
			err = keepers.CWErrorsKeeper.SetParams(unwrappedCtx, cwerrorsParams)
			if err != nil {
				return nil, err
			}

			// Setting cwica params
			cwicaParams, err := keepers.CWICAKeeper.GetParams(unwrappedCtx)
			if err != nil {
				return nil, err
			}
			cwicaParams.MsgSendTxMaxMessages = 5
			err = keepers.CWICAKeeper.SetParams(unwrappedCtx, cwicaParams)
			if err != nil {
				return nil, err
			}

			unwrappedCtx.Logger().Info(upgrades.ArchwayLogo + NameAsciiArt)
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
