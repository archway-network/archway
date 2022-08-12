package rewards

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/rewards/keeper"
	"github.com/archway-network/archway/x/rewards/types"
)

// BeginBlocker calculates and distributes dApp rewards for the previous block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	curBlockHeight := ctx.BlockHeight()
	if curBlockHeight <= 1 {
		return
	}

	k.CalculateRewards(ctx, curBlockHeight-1)
}
