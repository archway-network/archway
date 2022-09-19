package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/rewards/types"
)

// RegisterInvariants registers all module invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "module-account-balance", ModuleAccountBalanceInvariant(k))
}

// ModuleAccountBalanceInvariant checks that the current ModuleAccount pool funds are GTE type.RewardsRecord entries.
// If that one fails, calculated and stored rewards records are not "supported" by real tokens.
func ModuleAccountBalanceInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		poolCurrent := k.UndistributedRewardsPool(ctx)

		poolExpected := sdk.NewCoins()
		_, records := k.state.RewardsRecord(ctx).Export()
		for _, record := range records {
			poolExpected = poolExpected.Add(record.Rewards...)
		}

		broken := !poolExpected.IsEqual(poolCurrent)

		return sdk.FormatInvariant(types.ModuleName, "module account and total rewards records coins", fmt.Sprintf(
			"\tPool's tokens: %v\n"+
				"\tSum of rewards records tokens expected: %v\n"+
				"\tHeight: %d\n",
			poolCurrent, poolExpected, ctx.BlockHeight()),
		), broken
	}
}
