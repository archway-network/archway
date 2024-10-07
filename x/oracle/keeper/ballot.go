package keeper

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/math"

	"github.com/archway-network/archway/x/common/asset"
	"github.com/archway-network/archway/x/common/omap"
	"github.com/archway-network/archway/x/common/set"
	"github.com/archway-network/archway/x/oracle/types"
)

// GroupVotesByPair takes a collection of votes and groups them by their
// associated asset pair. This method only considers votes from active validators
// and disregards votes from validators that are not in the provided validator set.
//
// Note that any abstain votes (votes with a non-positive exchange rate) are
// assigned zero vote power. This function then returns a map where each
// asset pair is associated with its collection of ExchangeRateVotes.
func (k Keeper) GroupVotesByPair(
	ctx sdk.Context,
	validatorPerformances types.ValidatorPerformances,
) (pairVotes map[asset.Pair]types.ExchangeRateVotes) {
	pairVotes = map[asset.Pair]types.ExchangeRateVotes{}

	err := k.Votes.Walk(ctx, nil, func(voterAddrBytes []byte, aggregateVote types.AggregateExchangeRateVote) (bool, error) {
		voterAddr := sdk.ValAddress(voterAddrBytes)
		// skip votes from inactive validators
		validatorPerformance, exists := validatorPerformances[aggregateVote.Voter]
		if !exists {
			return false, nil
		}

		for _, tuple := range aggregateVote.ExchangeRateTuples {
			power := validatorPerformance.Power
			if !tuple.ExchangeRate.IsPositive() {
				// Make the power of abstain vote zero
				power = 0
			}

			pairVotes[tuple.Pair] = append(
				pairVotes[tuple.Pair],
				types.NewExchangeRateVote(
					tuple.ExchangeRate,
					tuple.Pair,
					voterAddr,
					power,
				),
			)
		}

		return false, nil
	})
	if err != nil {
		panic(err)
	}

	return
}

// ClearVotesAndPrevotes clears all tallied prevotes and votes from the store
func (k Keeper) ClearVotesAndPrevotes(ctx sdk.Context, votePeriod uint64) {
	// Clear all aggregate prevotes
	k.Prevotes.Walk(ctx, nil, func(valAddrBytes []byte, aggregatePrevote types.AggregateExchangeRatePrevote) (bool, error) {
		valAddr := sdk.ValAddress(valAddrBytes)
		if ctx.BlockHeight() >= int64(aggregatePrevote.SubmitBlock+votePeriod) {
			err := k.Prevotes.Remove(ctx, valAddr)
			if err != nil {
				k.Logger(ctx).Error("failed to delete prevote", "error", err)
			}
		}
		return false, nil
	})

	// Clear all aggregate votes
	iter, err := k.Votes.Iterate(ctx, nil)
	if err != nil {
		k.Logger(ctx).Error("failed to get votes iterator", "error", err)
		return
	}
	keys, err := iter.Keys()
	if err != nil {
		k.Logger(ctx).Error("failed to get keys for votes iterator", "error", err)
		return
	}
	for _, valAddr := range keys {
		err := k.Votes.Remove(ctx, valAddr)
		if err != nil {
			k.Logger(ctx).Error("failed to delete vote", "error", err)
		}
	}
}

// IsPassingVoteThreshold votes is passing the threshold amount of voting power
func IsPassingVoteThreshold(
	votes types.ExchangeRateVotes, thresholdVotingPower sdkmath.Int, minVoters uint64,
) bool {
	totalPower := math.NewInt(votes.Power())
	if totalPower.IsZero() {
		return false
	}

	if totalPower.LT(thresholdVotingPower) {
		return false
	}

	if votes.NumValidVoters() < minVoters {
		return false
	}

	return true
}

// RemoveInvalidVotes removes the votes which have not reached the vote
// threshold or which are not part of the whitelisted pairs anymore: example
// when params change during a vote period but some votes were already made.
//
// ALERT: This function mutates the pairVotes map, it removes the votes for
// the pair which is not passing the threshold or which is not whitelisted
// anymore.
func (k Keeper) RemoveInvalidVotes(
	ctx sdk.Context,
	pairVotes map[asset.Pair]types.ExchangeRateVotes,
	whitelistedPairs set.Set[asset.Pair],
) {
	boundTokens, err := k.StakingKeeper.TotalBondedTokens(ctx)
	if err != nil {
		panic(err)
	}
	totalBondedPower := sdk.TokensToConsensusPower(
		boundTokens,
		k.StakingKeeper.PowerReduction(ctx),
	)

	// Iterate through sorted keys for deterministic ordering.
	orderedPairVotes := omap.OrderedMap_Pair[types.ExchangeRateVotes](pairVotes)
	for pair := range orderedPairVotes.Range() {
		// If pair is not whitelisted, or the votes for it has failed, then skip
		// and remove it from pairBallotsMap for iteration efficiency
		if !whitelistedPairs.Has(pair) {
			delete(pairVotes, pair)
		}

		// If the votes is not passed, remove it from the whitelistedPairs set
		// to prevent slashing validators who did valid vote.
		if !IsPassingVoteThreshold(
			pairVotes[pair],
			k.VoteThreshold(ctx).MulInt64(totalBondedPower).RoundInt(),
			k.MinVoters(ctx),
		) {
			delete(whitelistedPairs, pair)
			delete(pairVotes, pair)
			continue
		}
	}
}

// Tally calculates the median and returns it. Sets the set of voters to be
// rewarded, i.e. voted within a reasonable spread from the weighted median to
// the store.
//
// ALERT: This function mutates validatorPerformances slice based on the votes
// made by the validators.
func Tally(
	votes types.ExchangeRateVotes,
	rewardBand math.LegacyDec,
	validatorPerformances types.ValidatorPerformances,
) math.LegacyDec {
	weightedMedian := votes.WeightedMedianWithAssertion()
	standardDeviation := votes.StandardDeviation(weightedMedian)
	rewardSpread := weightedMedian.Mul(rewardBand.QuoInt64(2))

	if standardDeviation.GT(rewardSpread) {
		rewardSpread = standardDeviation
	}

	for _, v := range votes {
		// Filter votes winners & abstain voters
		isInsideSpread := v.ExchangeRate.GTE(weightedMedian.Sub(rewardSpread)) &&
			v.ExchangeRate.LTE(weightedMedian.Add(rewardSpread))
		isAbstainVote := !v.ExchangeRate.IsPositive() // strictly less than zero, don't want to include zero
		isMiss := !isInsideSpread && !isAbstainVote

		validatorPerformance := validatorPerformances[v.Voter.String()]

		switch {
		case isInsideSpread:
			validatorPerformance.RewardWeight += v.Power
			validatorPerformance.WinCount++
		case isMiss:
			validatorPerformance.MissCount++
		case isAbstainVote:
			validatorPerformance.AbstainCount++
		}

		validatorPerformances[v.Voter.String()] = validatorPerformance
	}

	return weightedMedian
}
