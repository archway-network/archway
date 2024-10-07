package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SlashAndResetMissCounters do slash any operator who over criteria & clear all operators miss counter to zero
func (k Keeper) SlashAndResetMissCounters(ctx sdk.Context) {
	height := ctx.BlockHeight()
	distributionHeight := height - sdk.ValidatorUpdateDelay - 1

	// slash_window / vote_period
	votePeriodsPerWindow := uint64(
		math.LegacyNewDec(int64(k.SlashWindow(ctx))).
			QuoInt64(int64(k.VotePeriod(ctx))).
			TruncateInt64(),
	)
	minValidPerWindow := k.MinValidPerWindow(ctx)
	slashFraction := k.SlashFraction(ctx)
	powerReduction := k.StakingKeeper.PowerReduction(ctx)

	k.MissCounters.Walk(ctx, nil, func(operatorBytes []byte, missCounter uint64) (bool, error) {
		operator := sdk.ValAddress(operatorBytes)

		// Calculate valid vote rate; (SlashWindow - MissCounter)/SlashWindow
		validVoteRate := math.LegacyNewDecFromInt(
			math.NewInt(int64(votePeriodsPerWindow - missCounter))).
			QuoInt64(int64(votePeriodsPerWindow))

		// Penalize the validator whose the valid vote rate is smaller than min threshold
		if validVoteRate.LT(minValidPerWindow) {
			validator, err := k.StakingKeeper.Validator(ctx, operator)
			if err != nil {
				k.Logger(ctx).Error("fail to get validator", "operator", operator)
			}
			if validator.IsBonded() && !validator.IsJailed() {
				consAddr, err := validator.GetConsAddr()
				if err != nil {
					k.Logger(ctx).Error("fail to get consensus address", "validator", validator.GetOperator())
					return false, nil
				}

				k.slashingKeeper.Slash(
					ctx, consAddr, slashFraction, validator.GetConsensusPower(powerReduction), distributionHeight,
				)
				k.Logger(ctx).Info("oracle slash", "validator", string(consAddr), "fraction", slashFraction.String())
				k.slashingKeeper.Jail(ctx, consAddr)
			}
		}

		err := k.MissCounters.Remove(ctx, operator)
		if err != nil {
			k.Logger(ctx).Error("fail to delete miss counter", "operator", operator.String(), "error", err)
		}
		return false, nil
	})
}
