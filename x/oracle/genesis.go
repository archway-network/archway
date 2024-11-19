package oracle

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/oracle/asset"
	"github.com/archway-network/archway/x/oracle/keeper"
	"github.com/archway-network/archway/x/oracle/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, data *types.GenesisState) {
	for _, d := range data.FeederDelegations {
		voter, err := sdk.ValAddressFromBech32(d.ValidatorAddress)
		if err != nil {
			panic(err)
		}

		feeder, err := sdk.AccAddressFromBech32(d.FeederAddress)
		if err != nil {
			panic(err)
		}

		err = keeper.FeederDelegations.Set(ctx, voter, feeder)
		if err != nil {
			panic(err)
		}
	}

	for _, ex := range data.ExchangeRates {
		keeper.SetPrice(ctx, ex.Pair, ex.ExchangeRate)
	}

	for _, missCounter := range data.MissCounters {
		operator, err := sdk.ValAddressFromBech32(missCounter.ValidatorAddress)
		if err != nil {
			panic(err)
		}

		err = keeper.MissCounters.Set(ctx, operator, missCounter.MissCounter)
		if err != nil {
			panic(err)
		}
	}

	for _, aggregatePrevote := range data.AggregateExchangeRatePrevotes {
		valAddr, err := sdk.ValAddressFromBech32(aggregatePrevote.Voter)
		if err != nil {
			panic(err)
		}

		err = keeper.Prevotes.Set(ctx, valAddr, aggregatePrevote)
		if err != nil {
			panic(err)
		}
	}

	for _, aggregateVote := range data.AggregateExchangeRateVotes {
		valAddr, err := sdk.ValAddressFromBech32(aggregateVote.Voter)
		if err != nil {
			panic(err)
		}

		err = keeper.Votes.Set(ctx, valAddr, aggregateVote)
		if err != nil {
			panic(err)
		}
	}

	if len(data.Pairs) > 0 {
		for _, tt := range data.Pairs {
			err := keeper.WhitelistedPairs.Set(ctx, tt)
			if err != nil {
				panic(err)
			}
		}
	} else {
		for _, item := range data.Params.Whitelist {
			err := keeper.WhitelistedPairs.Set(ctx, item)
			if err != nil {
				panic(err)
			}
		}
	}

	for _, pr := range data.Rewards {
		err := keeper.Rewards.Set(ctx, pr.Id, pr)
		if err != nil {
			panic(err)
		}
	}

	// set last ID based on the last pair reward
	if len(data.Rewards) != 0 {
		err := keeper.RewardsID.Set(ctx, data.Rewards[len(data.Rewards)-1].Id)
		if err != nil {
			panic(err)
		}
	}
	err := keeper.Params.Set(ctx, data.Params)
	if err != nil {
		panic(err)
	}

	// check if the module account exists
	moduleAcc := keeper.AccountKeeper.GetModuleAccount(ctx, types.ModuleName)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	params, err := keeper.Params.Get(ctx)
	if err != nil {
		panic(err)
	}

	feederDelegations := []types.FeederDelegation{}
	err = keeper.FeederDelegations.Walk(ctx, nil, func(valBytes []byte, accBytes []byte) (bool, error) {
		feederDelegations = append(feederDelegations, types.FeederDelegation{
			FeederAddress:    sdk.AccAddress(accBytes).String(),
			ValidatorAddress: sdk.ValAddress(valBytes).String(),
		})
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	exchangeRates := []types.ExchangeRateTuple{}
	err = keeper.ExchangeRates.Walk(ctx, nil, func(pair asset.Pair, price types.DatedPrice) (bool, error) {
		exchangeRates = append(exchangeRates, types.ExchangeRateTuple{
			Pair:         pair,
			ExchangeRate: price.ExchangeRate,
		})
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	missCounters := []types.MissCounter{}
	err = keeper.MissCounters.Walk(ctx, nil, func(valAddrBytes []byte, counter uint64) (bool, error) {
		missCounters = append(missCounters, types.MissCounter{
			ValidatorAddress: sdk.ValAddress(valAddrBytes).String(),
			MissCounter:      counter,
		})
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	var pairs []asset.Pair
	iter, err := keeper.WhitelistedPairs.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	keys, err := iter.Keys()
	if err != nil {
		panic(err)
	}
	pairs = append(pairs, keys...)

	prevotesIter, err := keeper.Prevotes.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	prevotes, err := prevotesIter.Values()
	if err != nil {
		panic(err)
	}

	votesIter, err := keeper.Votes.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	votes, err := votesIter.Values()
	if err != nil {
		panic(err)
	}

	rewardsIter, err := keeper.Rewards.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	rewards, err := rewardsIter.Values()
	if err != nil {
		panic(err)
	}

	return types.NewGenesisState(
		params,
		exchangeRates,
		feederDelegations,
		missCounters,
		prevotes,
		votes,
		pairs,
		rewards,
	)
}
