package rewards

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/archway-network/archway/x/rewards/keeper"
	"github.com/archway-network/archway/x/rewards/types"
)

// EndBlocker calculates and distributes dApp rewards for the current block updating the treasury.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	k.AllocateBlockRewards(ctx, ctx.BlockHeight())

	return []abci.ValidatorUpdate{}
}
