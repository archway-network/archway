package oracle_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/oracle"
	"github.com/archway-network/archway/x/oracle/types"
)

func TestExportInitGenesis(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1, e2eTesting.WithValidatorsNum(2))
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()

	vals := chain.GetCurrentValSet().Validators
	AccAddrs := make([]sdk.AccAddress, len(vals))
	ValAddrs := make([]sdk.ValAddress, len(vals))
	for i := range vals {
		AccAddrs[i] = sdk.AccAddress(vals[i].Address)
		ValAddrs[i] = sdk.ValAddress(vals[i].Address)
	}

	require.NoError(t, keepers.OracleKeeper.Params.Set(ctx, types.DefaultParams()))
	require.NoError(t, keepers.OracleKeeper.FeederDelegations.Set(ctx, ValAddrs[0], AccAddrs[1]))
	require.NoError(t, keepers.OracleKeeper.ExchangeRates.Set(ctx, "pair1:pair2", types.DatedPrice{
		ExchangeRate:       math.LegacyNewDec(123),
		CreationHeight:     0,
		CreationTimeUnixMs: 0,
	}))
	require.NoError(t, keepers.OracleKeeper.Prevotes.Set(ctx, ValAddrs[0], types.NewAggregateExchangeRatePrevote(types.AggregateVoteHash{123}, ValAddrs[0], uint64(2))))
	require.NoError(t, keepers.OracleKeeper.Votes.Set(ctx, ValAddrs[0], types.NewAggregateExchangeRateVote(types.ExchangeRateTuples{{Pair: "foo", ExchangeRate: math.LegacyNewDec(123)}}, ValAddrs[0])))
	require.NoError(t, keepers.OracleKeeper.WhitelistedPairs.Set(ctx, "pair1:pair1"))
	require.NoError(t, keepers.OracleKeeper.WhitelistedPairs.Set(ctx, "pair2:pair2"))
	require.NoError(t, keepers.OracleKeeper.MissCounters.Set(ctx, ValAddrs[0], 10))
	require.NoError(t, keepers.OracleKeeper.Rewards.Set(ctx, 0, types.Rewards{
		Id:          0,
		VotePeriods: 100,
		Coins:       sdk.NewCoins(sdk.NewInt64Coin("test", 1000)),
	}))
	genesis := oracle.ExportGenesis(ctx, keepers.OracleKeeper)

	chain = e2eTesting.NewTestChain(t, 2)
	keepers = chain.GetApp().Keepers
	ctx = chain.GetContext()
	oracle.InitGenesis(ctx, keepers.OracleKeeper, genesis)
	newGenesis := oracle.ExportGenesis(ctx, keepers.OracleKeeper)

	require.Equal(t, genesis, newGenesis)
}

func TestInitGenesis(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1, e2eTesting.WithValidatorsNum(1))
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()

	vals := chain.GetCurrentValSet().Validators
	AccAddrs := make([]sdk.AccAddress, len(vals))
	ValAddrs := make([]sdk.ValAddress, len(vals))
	for i := range vals {
		AccAddrs[i] = sdk.AccAddress(vals[i].Address)
		ValAddrs[i] = sdk.ValAddress(vals[i].Address)
	}

	genesis := types.DefaultGenesisState()
	require.NotPanics(t, func() {
		oracle.InitGenesis(ctx, keepers.OracleKeeper, genesis)
	})

	genesis.FeederDelegations = []types.FeederDelegation{{
		FeederAddress:    AccAddrs[0].String(),
		ValidatorAddress: "invalid",
	}}

	require.Panics(t, func() {
		oracle.InitGenesis(ctx, keepers.OracleKeeper, genesis)
	})

	genesis.FeederDelegations = []types.FeederDelegation{{
		FeederAddress:    "invalid",
		ValidatorAddress: ValAddrs[0].String(),
	}}

	require.Panics(t, func() {
		oracle.InitGenesis(ctx, keepers.OracleKeeper, genesis)
	})

	genesis.FeederDelegations = []types.FeederDelegation{{
		FeederAddress:    AccAddrs[0].String(),
		ValidatorAddress: ValAddrs[0].String(),
	}}

	genesis.MissCounters = []types.MissCounter{
		{
			ValidatorAddress: "invalid",
			MissCounter:      10,
		},
	}

	require.Panics(t, func() {
		oracle.InitGenesis(ctx, keepers.OracleKeeper, genesis)
	})

	genesis.MissCounters = []types.MissCounter{
		{
			ValidatorAddress: ValAddrs[0].String(),
			MissCounter:      10,
		},
	}

	genesis.AggregateExchangeRatePrevotes = []types.AggregateExchangeRatePrevote{
		{
			Hash:        "hash",
			Voter:       "invalid",
			SubmitBlock: 100,
		},
	}

	require.Panics(t, func() {
		oracle.InitGenesis(ctx, keepers.OracleKeeper, genesis)
	})

	genesis.AggregateExchangeRatePrevotes = []types.AggregateExchangeRatePrevote{
		{
			Hash:        "hash",
			Voter:       ValAddrs[0].String(),
			SubmitBlock: 100,
		},
	}

	genesis.AggregateExchangeRateVotes = []types.AggregateExchangeRateVote{
		{
			ExchangeRateTuples: []types.ExchangeRateTuple{
				{
					Pair:         "nibi:usd",
					ExchangeRate: math.LegacyNewDec(10),
				},
			},
			Voter: "invalid",
		},
	}

	require.Panics(t, func() {
		oracle.InitGenesis(ctx, keepers.OracleKeeper, genesis)
	})

	genesis.AggregateExchangeRateVotes = []types.AggregateExchangeRateVote{
		{
			ExchangeRateTuples: []types.ExchangeRateTuple{
				{
					Pair:         "nibi:usd",
					ExchangeRate: math.LegacyNewDec(10),
				},
			},
			Voter: ValAddrs[0].String(),
		},
	}

	require.NotPanics(t, func() {
		oracle.InitGenesis(ctx, keepers.OracleKeeper, genesis)
	})
}
