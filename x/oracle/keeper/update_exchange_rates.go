package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/types/set"
	"github.com/archway-network/archway/x/oracle/asset"
	"github.com/archway-network/archway/x/oracle/types"
)

// UpdateExchangeRates updates the ExchangeRates, this is supposed to be executed on EndBlock.
func (k Keeper) UpdateExchangeRates(ctx sdk.Context) types.ValidatorPerformances {
	k.Logger(ctx).Info("processing validator price votes")
	validatorPerformances := k.newValidatorPerformances(ctx)
	pairs, err := k.GetWhitelistedPairs(ctx)
	if err != nil {
		panic(err)
	}
	whitelistedPairs := set.New[asset.Pair](pairs...)

	pairVotes := k.getPairVotes(ctx, validatorPerformances, whitelistedPairs)

	k.ClearExchangeRates(ctx, pairVotes)
	k.tallyVotesAndUpdatePrices(ctx, pairVotes, validatorPerformances)

	k.incrementMissCounters(ctx, validatorPerformances)
	k.incrementAbstainsByOmission(ctx, whitelistedPairs.Len(), validatorPerformances)

	k.rewardWinners(ctx, validatorPerformances)

	params, _ := k.Params.Get(ctx)
	k.ClearVotesAndPrevotes(ctx, params.VotePeriod)
	k.RefreshWhitelist(ctx, params.Whitelist, whitelistedPairs)

	for _, validatorPerformance := range validatorPerformances {
		_ = ctx.EventManager().EmitTypedEvent(&types.EventValidatorPerformance{
			Validator:    validatorPerformance.ValAddress.String(),
			VotingPower:  validatorPerformance.Power,
			RewardWeight: validatorPerformance.RewardWeight,
			WinCount:     validatorPerformance.WinCount,
			AbstainCount: validatorPerformance.AbstainCount,
			MissCount:    validatorPerformance.MissCount,
		})
	}

	return validatorPerformances
}

// incrementMissCounters it parses all validators performance and increases the
// missed vote of those that did not vote.
func (k Keeper) incrementMissCounters(
	ctx sdk.Context,
	validatorPerformances types.ValidatorPerformances,
) {
	for _, validatorPerformance := range validatorPerformances {
		if int(validatorPerformance.MissCount) > 0 {
			counter, err := k.MissCounters.Get(ctx, validatorPerformance.ValAddress)
			if err != nil {
				counter = 0
			}
			counter = counter + uint64(validatorPerformance.MissCount)
			err = k.MissCounters.Set(
				ctx, validatorPerformance.ValAddress,
				counter,
			)
			if err == nil {
				k.Logger(ctx).Info("vote miss", "validator", validatorPerformance.ValAddress.String())
			} else {
				k.Logger(ctx).Error("failed to set MissCounter", "validator", validatorPerformance.ValAddress.String(), "counter", counter, "error", err)
			}
		}
	}
}

func (k Keeper) incrementAbstainsByOmission(
	_ sdk.Context,
	numPairs int,
	validatorPerformances types.ValidatorPerformances,
) {
	for valAddr, performance := range validatorPerformances {
		omitCount := int64(numPairs) - (performance.WinCount + performance.AbstainCount + performance.MissCount)
		if omitCount > 0 {
			performance.AbstainCount += omitCount
			validatorPerformances[valAddr] = performance
		}
	}
}

// tallyVotesAndUpdatePrices processes the votes and updates the ExchangeRates based on the results.
func (k Keeper) tallyVotesAndUpdatePrices(
	ctx sdk.Context,
	pairVotes map[asset.Pair]types.ExchangeRateVotes,
	validatorPerformances types.ValidatorPerformances,
) {
	rewardBand := k.RewardBand(ctx)
	for pair, votes := range pairVotes {
		exchangeRate := Tally(votes, rewardBand, validatorPerformances)
		k.SetPrice(ctx, pair, exchangeRate)
	}
}

// getPairVotes returns a map of pairs and votes excluding abstained votes and votes that don't meet the threshold criteria
func (k Keeper) getPairVotes(
	ctx sdk.Context,
	validatorPerformances types.ValidatorPerformances,
	whitelistedPairs set.Set[asset.Pair],
) map[asset.Pair]types.ExchangeRateVotes {
	pairVotes := k.GroupVotesByPair(ctx, validatorPerformances)

	k.RemoveInvalidVotes(ctx, pairVotes, whitelistedPairs)

	return pairVotes
}

// ClearExchangeRates removes all exchange rates from the state
// We remove the price for pair with expired prices or valid votes
func (k Keeper) ClearExchangeRates(ctx sdk.Context, pairVotes map[asset.Pair]types.ExchangeRateVotes) {
	params, _ := k.Params.Get(ctx)

	_ = k.ExchangeRates.Walk(ctx, nil, func(key asset.Pair, _ types.DatedPrice) (bool, error) {
		_, isValid := pairVotes[key]
		previousExchangeRate, _ := k.ExchangeRates.Get(ctx, key)
		isExpired := previousExchangeRate.CreatedBlock+params.ExpirationBlocks <= uint64(ctx.BlockHeight())

		if isValid || isExpired {
			err := k.ExchangeRates.Remove(ctx, key)
			if err != nil {
				k.Logger(ctx).Error("failed to delete exchange rate", "pair", key.String(), "error", err)
			}
		}
		return false, nil
	})
}

// newValidatorPerformances creates a new map of validators and their performance, excluding validators that are
// not bonded.
func (k Keeper) newValidatorPerformances(ctx sdk.Context) types.ValidatorPerformances {
	validatorPerformances := make(map[string]types.ValidatorPerformance)

	maxValidators, err := k.StakingKeeper.MaxValidators(ctx)
	if err != nil {
		panic(err)
	}
	powerReduction := k.StakingKeeper.PowerReduction(ctx)

	iterator, err := k.StakingKeeper.ValidatorsPowerStoreIterator(ctx)
	if err != nil {
		panic(err)
	}
	defer iterator.Close()

	for i := 0; iterator.Valid() && i < int(maxValidators); iterator.Next() {
		validator, err := k.StakingKeeper.Validator(ctx, iterator.Value())

		// exclude not bonded
		if err != nil || !validator.IsBonded() {
			continue
		}

		valAddrStr := validator.GetOperator()
		valAddr, err := sdk.ValAddressFromBech32(valAddrStr)
		if err != nil {
			panic(err)
		}
		validatorPerformances[valAddrStr] = types.NewValidatorPerformance(
			validator.GetConsensusPower(powerReduction),
			valAddr,
		)
		i++
	}

	return validatorPerformances
}
