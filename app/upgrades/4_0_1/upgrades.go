package upgradelatest

import (
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/archway-network/archway/app/upgrades"
)

const Name = "v4.0.1"

var Upgrade = upgrades.Upgrade{
	UpgradeName: Name,
	CreateUpgradeHandler: func(mm *module.Manager, cfg module.Configurator, accountKeeper keeper.AccountKeeper) upgradetypes.UpgradeHandler {
		return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			fcAccount := accountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
			account, ok := fcAccount.(*authtypes.ModuleAccount)
			if !ok {
				return nil, fmt.Errorf("feeCollector account is not *authtypes.ModuleAccount")
			}
			account.Permissions = append(account.Permissions, authtypes.Burner)
			accountKeeper.SetModuleAccount(ctx, account)

			return mm.RunMigrations(ctx, cfg, fromVM)
		}
	},
	StoreUpgrades: storetypes.StoreUpgrades{},
}
