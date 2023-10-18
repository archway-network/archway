package keeper

import (
	"math"

	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TrackFeeRebatesRewards creates a new transaction fee rebate reward record for the current transaction.
// Unique transaction ID is taken from the tracking module.
// CONTRACT: tracking Ante handler must be called before this module's Ante handler (tracking provides the primary key).
func (k Keeper) TrackFeeRebatesRewards(ctx sdk.Context, rewards sdk.Coins) {
	txID := k.trackingKeeper.GetCurrentTxID(ctx)
	err := k.TxRewards.Set(ctx, txID, rewardsTypes.TxRewards{
		TxId:       txID,
		Height:     ctx.BlockHeight(),
		FeeRewards: rewards,
	})
	if err != nil {
		panic(err)
	}
}

// TrackInflationRewards creates a new inflation reward record for the current block.
func (k Keeper) TrackInflationRewards(ctx sdk.Context, rewards sdk.Coin) {
	blockGasLimit := ctx.BlockGasMeter().Limit()
	if ctx.BlockGasMeter().Limit() == math.MaxUint64 { // Because thisss https://github.com/cosmos/cosmos-sdk/pull/9651
		blockGasLimit = 0
	}

	err := k.BlockRewards.Set(ctx, uint64(ctx.BlockHeight()), rewardsTypes.BlockRewards{
		Height:           ctx.BlockHeight(),
		InflationRewards: rewards,
		MaxGas:           blockGasLimit,
	})
	if err != nil {
		panic(err)
	}
}
