package upgrade4_0_2

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/archway-network/archway/app/keepers"
	"github.com/archway-network/archway/app/upgrades"
)

const Name = "v4.0.2"

var Upgrade = upgrades.Upgrade{
	UpgradeName: Name,
	CreateUpgradeHandler: func(mm *module.Manager, cfg module.Configurator, keepers keepers.ArchwayKeepers) upgradetypes.UpgradeHandler {
		return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			accountKeeper := keepers.AccountKeeper
			fcAccount := accountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
			account, ok := fcAccount.(*authtypes.ModuleAccount)
			if !ok {
				return nil, fmt.Errorf("feeCollector account is not *authtypes.ModuleAccount")
			}
			if !account.HasPermission(authtypes.Burner) {
				account.Permissions = append(account.Permissions, authtypes.Burner)
			}
			err := accountKeeper.ValidatePermissions(account)
			if err != nil {
				return nil, fmt.Errorf("Could not validate feeCollectors permissions")
			}
			accountKeeper.SetModuleAccount(ctx, account)

			return mm.RunMigrations(ctx, cfg, fromVM)
		}
	},
	StoreUpgrades: storetypes.StoreUpgrades{},
}
