package keeper_test

import (
	"sort"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/NibiruChain/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	testutilevents "github.com/archway-network/archway/x/common/testutil"

	"github.com/archway-network/archway/x/common/asset"
	"github.com/archway-network/archway/x/common/denoms"
	"github.com/archway-network/archway/x/oracle/keeper"
	"github.com/archway-network/archway/x/oracle/types"
)

func TestQueryParams(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()

	querier := keeper.NewQuerier(keepers.OracleKeeper)
	res, err := querier.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)

	params, err := keepers.OracleKeeper.Params.Get(ctx)
	require.NoError(t, err)

	require.Equal(t, params, res.Params)
}

func TestQueryExchangeRate(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()
	querier := keeper.NewQuerier(keepers.OracleKeeper)

	rate := math.LegacyNewDec(1700)
	keepers.OracleKeeper.ExchangeRates.Insert(ctx, asset.Registry.Pair(denoms.ETH, denoms.NUSD), types.DatedPrice{ExchangeRate: rate, CreatedBlock: uint64(ctx.BlockHeight())})

	// empty request
	_, err := querier.ExchangeRate(ctx, nil)
	require.Error(t, err)

	// Query to grpc
	res, err := querier.ExchangeRate(ctx, &types.QueryExchangeRateRequest{
		Pair: asset.Registry.Pair(denoms.ETH, denoms.NUSD),
	})
	require.NoError(t, err)
	require.Equal(t, rate, res.ExchangeRate)
}

func TestQueryMissCounter(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1, e2eTesting.WithValidatorsNum(1))
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()
	querier := keeper.NewQuerier(keepers.OracleKeeper)

	vals := chain.GetCurrentValSet().Validators
	ValAddrs := make([]sdk.ValAddress, len(vals))
	for i := range vals {
		ValAddrs[i] = sdk.ValAddress(vals[i].Address)
	}

	missCounter := uint64(1)
	keepers.OracleKeeper.MissCounters.Insert(ctx, ValAddrs[0], missCounter)

	// empty request
	_, err := querier.MissCounter(ctx, nil)
	require.Error(t, err)

	// Query to grpc
	res, err := querier.MissCounter(ctx, &types.QueryMissCounterRequest{
		ValidatorAddr: ValAddrs[0].String(),
	})
	require.NoError(t, err)
	require.Equal(t, missCounter, res.MissCounter)
}

func TestQueryExchangeRates(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()
	querier := keeper.NewQuerier(keepers.OracleKeeper)

	rate := math.LegacyNewDec(1700)
	keepers.OracleKeeper.ExchangeRates.Insert(ctx, asset.Registry.Pair(denoms.BTC, denoms.NUSD), types.DatedPrice{ExchangeRate: rate, CreatedBlock: uint64(ctx.BlockHeight())})
	keepers.OracleKeeper.ExchangeRates.Insert(ctx, asset.Registry.Pair(denoms.ETH, denoms.NUSD), types.DatedPrice{ExchangeRate: rate, CreatedBlock: uint64(ctx.BlockHeight())})

	res, err := querier.ExchangeRates(ctx, &types.QueryExchangeRatesRequest{})
	require.NoError(t, err)

	require.Equal(t, types.ExchangeRateTuples{
		{Pair: asset.Registry.Pair(denoms.BTC, denoms.NUSD), ExchangeRate: rate},
		{Pair: asset.Registry.Pair(denoms.ETH, denoms.NUSD), ExchangeRate: rate},
	}, res.ExchangeRates)
}

func TestQueryExchangeRateTwap(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()
	querier := keeper.NewQuerier(keepers.OracleKeeper)

	rate := math.LegacyNewDec(1700)
	keepers.OracleKeeper.SetPrice(ctx, asset.Registry.Pair(denoms.BTC, denoms.NUSD), rate)
	testutilevents.RequireContainsTypedEvent(
		t,
		ctx,
		&types.EventPriceUpdate{
			Pair:        asset.Registry.Pair(denoms.BTC, denoms.NUSD).String(),
			Price:       rate,
			TimestampMs: ctx.BlockTime().UnixMilli(),
		},
	)

	ctx = ctx.
		WithBlockTime(ctx.BlockTime().Add(time.Second)).
		WithBlockHeight(ctx.BlockHeight() + 1)

	_, err := querier.ExchangeRateTwap(ctx, &types.QueryExchangeRateRequest{Pair: asset.Registry.Pair(denoms.ETH, denoms.NUSD)})
	require.Error(t, err)

	res, err := querier.ExchangeRateTwap(ctx, &types.QueryExchangeRateRequest{Pair: asset.Registry.Pair(denoms.BTC, denoms.NUSD)})
	require.NoError(t, err)
	require.Equal(t, math.LegacyMustNewDecFromStr("1700"), res.ExchangeRate)
}

func TestCalcTwap(t *testing.T) {
	tests := []struct {
		name               string
		pair               asset.Pair
		priceSnapshots     []types.PriceSnapshot
		currentBlockTime   time.Time
		currentBlockHeight int64
		lookbackInterval   time.Duration
		assetAmount        math.LegacyDec
		expectedPrice      math.LegacyDec
		expectedErr        error
	}{
		// expected price: (9.5 * (35 - 30) + 8.5 * (30 - 20) + 9.0 * (20 - 5)) / 30 = 8.916666
		{
			name: "spot price twap calc, t=(5,35]",
			pair: asset.Registry.Pair(denoms.BTC, denoms.NUSD),
			priceSnapshots: []types.PriceSnapshot{
				{
					Pair:        asset.Registry.Pair(denoms.BTC, denoms.NUSD),
					Price:       math.LegacyMustNewDecFromStr("90000.0"),
					TimestampMs: time.UnixMilli(1).UnixMilli(),
				},
				{
					Pair:        asset.Registry.Pair(denoms.BTC, denoms.NUSD),
					Price:       math.LegacyMustNewDecFromStr("9.0"),
					TimestampMs: time.UnixMilli(10).UnixMilli(),
				},
				{
					Pair:        asset.Registry.Pair(denoms.BTC, denoms.NUSD),
					Price:       math.LegacyMustNewDecFromStr("8.5"),
					TimestampMs: time.UnixMilli(20).UnixMilli(),
				},
				{
					Pair:        asset.Registry.Pair(denoms.BTC, denoms.NUSD),
					Price:       math.LegacyMustNewDecFromStr("9.5"),
					TimestampMs: time.UnixMilli(30).UnixMilli(),
				},
			},
			currentBlockTime:   time.UnixMilli(35),
			currentBlockHeight: 3,
			lookbackInterval:   30 * time.Millisecond,
			expectedPrice:      math.LegacyMustNewDecFromStr("8.900000000000000000"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			chain := e2eTesting.NewTestChain(t, 1)
			keepers := chain.GetApp().Keepers
			querier := keeper.NewQuerier(keepers.OracleKeeper)
			ctx := chain.GetContext()

			newParams := types.Params{
				VotePeriod:         types.DefaultVotePeriod,
				VoteThreshold:      types.DefaultVoteThreshold,
				MinVoters:          types.DefaultMinVoters,
				RewardBand:         types.DefaultRewardBand,
				Whitelist:          types.DefaultWhitelist,
				SlashFraction:      types.DefaultSlashFraction,
				SlashWindow:        types.DefaultSlashWindow,
				MinValidPerWindow:  types.DefaultMinValidPerWindow,
				TwapLookbackWindow: tc.lookbackInterval,
				ValidatorFeeRatio:  types.DefaultValidatorFeeRatio,
			}

			keepers.OracleKeeper.Params.Set(ctx, newParams)
			ctx = ctx.WithBlockTime(time.UnixMilli(0))
			for _, reserve := range tc.priceSnapshots {
				ctx = ctx.WithBlockTime(time.UnixMilli(reserve.TimestampMs))
				keepers.OracleKeeper.SetPrice(ctx, asset.Registry.Pair(denoms.BTC, denoms.NUSD), reserve.Price)
			}

			ctx = ctx.WithBlockTime(tc.currentBlockTime).WithBlockHeight(tc.currentBlockHeight)

			price, err := querier.ExchangeRateTwap(sdk.WrapSDKContext(ctx), &types.QueryExchangeRateRequest{Pair: asset.Registry.Pair(denoms.BTC, denoms.NUSD)})
			require.NoError(t, err)

			require.EqualValuesf(t, tc.expectedPrice, price.ExchangeRate,
				"expected %s, got %s", tc.expectedPrice.String(), price.ExchangeRate.String())
		})
	}
}

func TestQueryActives(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()
	queryClient := keeper.NewQuerier(keepers.OracleKeeper)

	rate := math.LegacyNewDec(1700)
	keepers.OracleKeeper.ExchangeRates.Insert(ctx, asset.Registry.Pair(denoms.BTC, denoms.NUSD), types.DatedPrice{ExchangeRate: rate, CreatedBlock: uint64(ctx.BlockHeight())})
	keepers.OracleKeeper.ExchangeRates.Insert(ctx, asset.Registry.Pair(denoms.NIBI, denoms.NUSD), types.DatedPrice{ExchangeRate: rate, CreatedBlock: uint64(ctx.BlockHeight())})
	keepers.OracleKeeper.ExchangeRates.Insert(ctx, asset.Registry.Pair(denoms.ETH, denoms.NUSD), types.DatedPrice{ExchangeRate: rate, CreatedBlock: uint64(ctx.BlockHeight())})

	res, err := queryClient.Actives(ctx, &types.QueryActivesRequest{})
	require.NoError(t, err)

	targetPairs := []asset.Pair{
		asset.Registry.Pair(denoms.BTC, denoms.NUSD),
		asset.Registry.Pair(denoms.ETH, denoms.NUSD),
		asset.Registry.Pair(denoms.NIBI, denoms.NUSD),
	}

	require.Equal(t, targetPairs, res.Actives)
}

func TestQueryFeederDelegation(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1, e2eTesting.WithValidatorsNum(2))
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()
	querier := keeper.NewQuerier(keepers.OracleKeeper)

	vals := chain.GetCurrentValSet().Validators
	AccAddrs := make([]sdk.AccAddress, len(vals))
	ValAddrs := make([]sdk.ValAddress, len(vals))
	for i := range vals {
		AccAddrs[i] = sdk.AccAddress(vals[i].Address)
		ValAddrs[i] = sdk.ValAddress(vals[i].Address)
	}

	keepers.OracleKeeper.FeederDelegations.Insert(ctx, ValAddrs[0], AccAddrs[1])

	// empty request
	_, err := querier.FeederDelegation(ctx, nil)
	require.Error(t, err)

	res, err := querier.FeederDelegation(ctx, &types.QueryFeederDelegationRequest{
		ValidatorAddr: ValAddrs[0].String(),
	})
	require.NoError(t, err)

	require.Equal(t, AccAddrs[1].String(), res.FeederAddr)
}

func TestQueryAggregatePrevote(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1, e2eTesting.WithValidatorsNum(2))
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()
	querier := keeper.NewQuerier(keepers.OracleKeeper)

	vals := chain.GetCurrentValSet().Validators
	ValAddrs := make([]sdk.ValAddress, len(vals))
	for i := range vals {
		ValAddrs[i] = sdk.ValAddress(vals[i].Address)
	}

	prevote1 := types.NewAggregateExchangeRatePrevote(types.AggregateVoteHash{}, ValAddrs[0], 0)
	keepers.OracleKeeper.Prevotes.Insert(ctx, ValAddrs[0], prevote1)
	prevote2 := types.NewAggregateExchangeRatePrevote(types.AggregateVoteHash{}, ValAddrs[1], 0)
	keepers.OracleKeeper.Prevotes.Insert(ctx, ValAddrs[1], prevote2)

	// validator 0 address params
	res, err := querier.AggregatePrevote(ctx, &types.QueryAggregatePrevoteRequest{
		ValidatorAddr: ValAddrs[0].String(),
	})
	require.NoError(t, err)
	require.Equal(t, prevote1, res.AggregatePrevote)

	// empty request
	_, err = querier.AggregatePrevote(ctx, nil)
	require.Error(t, err)

	// validator 1 address params
	res, err = querier.AggregatePrevote(ctx, &types.QueryAggregatePrevoteRequest{
		ValidatorAddr: ValAddrs[1].String(),
	})
	require.NoError(t, err)
	require.Equal(t, prevote2, res.AggregatePrevote)
}

func TestQueryAggregatePrevotes(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1, e2eTesting.WithValidatorsNum(3))
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()
	querier := keeper.NewQuerier(keepers.OracleKeeper)

	vals := chain.GetCurrentValSet().Validators
	ValAddrs := make([]sdk.ValAddress, len(vals))
	for i := range vals {
		ValAddrs[i] = sdk.ValAddress(vals[i].Address)
	}

	prevote1 := types.NewAggregateExchangeRatePrevote(types.AggregateVoteHash{}, ValAddrs[0], 0)
	keepers.OracleKeeper.Prevotes.Insert(ctx, ValAddrs[0], prevote1)
	prevote2 := types.NewAggregateExchangeRatePrevote(types.AggregateVoteHash{}, ValAddrs[1], 0)
	keepers.OracleKeeper.Prevotes.Insert(ctx, ValAddrs[1], prevote2)
	prevote3 := types.NewAggregateExchangeRatePrevote(types.AggregateVoteHash{}, ValAddrs[2], 0)
	keepers.OracleKeeper.Prevotes.Insert(ctx, ValAddrs[2], prevote3)

	expectedPrevotes := []types.AggregateExchangeRatePrevote{prevote1, prevote2, prevote3}
	sort.SliceStable(expectedPrevotes, func(i, j int) bool {
		return expectedPrevotes[i].Voter <= expectedPrevotes[j].Voter
	})

	res, err := querier.AggregatePrevotes(ctx, &types.QueryAggregatePrevotesRequest{})
	require.NoError(t, err)
	require.Equal(t, expectedPrevotes, res.AggregatePrevotes)
}

func TestQueryAggregateVote(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1, e2eTesting.WithValidatorsNum(2))
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()
	querier := keeper.NewQuerier(keepers.OracleKeeper)

	vals := chain.GetCurrentValSet().Validators
	ValAddrs := make([]sdk.ValAddress, len(vals))
	for i := range vals {
		ValAddrs[i] = sdk.ValAddress(vals[i].Address)
	}

	vote1 := types.NewAggregateExchangeRateVote(types.ExchangeRateTuples{{Pair: "", ExchangeRate: math.LegacyOneDec()}}, ValAddrs[0])
	keepers.OracleKeeper.Votes.Insert(ctx, ValAddrs[0], vote1)
	vote2 := types.NewAggregateExchangeRateVote(types.ExchangeRateTuples{{Pair: "", ExchangeRate: math.LegacyOneDec()}}, ValAddrs[1])
	keepers.OracleKeeper.Votes.Insert(ctx, ValAddrs[1], vote2)

	// empty request
	_, err := querier.AggregateVote(ctx, nil)
	require.Error(t, err)

	// validator 0 address params
	res, err := querier.AggregateVote(ctx, &types.QueryAggregateVoteRequest{
		ValidatorAddr: ValAddrs[0].String(),
	})
	require.NoError(t, err)
	require.Equal(t, vote1, res.AggregateVote)

	// validator 1 address params
	res, err = querier.AggregateVote(ctx, &types.QueryAggregateVoteRequest{
		ValidatorAddr: ValAddrs[1].String(),
	})
	require.NoError(t, err)
	require.Equal(t, vote2, res.AggregateVote)
}

func TestQueryAggregateVotes(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1, e2eTesting.WithValidatorsNum(3))
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()
	querier := keeper.NewQuerier(keepers.OracleKeeper)

	vals := chain.GetCurrentValSet().Validators
	ValAddrs := make([]sdk.ValAddress, len(vals))
	for i := range vals {
		ValAddrs[i] = sdk.ValAddress(vals[i].Address)
	}

	vote1 := types.NewAggregateExchangeRateVote(types.ExchangeRateTuples{{Pair: "", ExchangeRate: math.LegacyOneDec()}}, ValAddrs[0])
	keepers.OracleKeeper.Votes.Insert(ctx, ValAddrs[0], vote1)
	vote2 := types.NewAggregateExchangeRateVote(types.ExchangeRateTuples{{Pair: "", ExchangeRate: math.LegacyOneDec()}}, ValAddrs[1])
	keepers.OracleKeeper.Votes.Insert(ctx, ValAddrs[1], vote2)
	vote3 := types.NewAggregateExchangeRateVote(types.ExchangeRateTuples{{Pair: "", ExchangeRate: math.LegacyOneDec()}}, ValAddrs[2])
	keepers.OracleKeeper.Votes.Insert(ctx, ValAddrs[2], vote3)

	expectedVotes := []types.AggregateExchangeRateVote{vote1, vote2, vote3}
	sort.SliceStable(expectedVotes, func(i, j int) bool {
		return expectedVotes[i].Voter <= expectedVotes[j].Voter
	})

	res, err := querier.AggregateVotes(ctx, &types.QueryAggregateVotesRequest{})
	require.NoError(t, err)
	require.Equal(t, expectedVotes, res.AggregateVotes)
}

func TestQueryVoteTargets(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()
	querier := keeper.NewQuerier(keepers.OracleKeeper)

	// clear pairs
	for _, p := range keepers.OracleKeeper.WhitelistedPairs.Iterate(ctx, collections.Range[asset.Pair]{}).Keys() {
		keepers.OracleKeeper.WhitelistedPairs.Delete(ctx, p)
	}

	voteTargets := []asset.Pair{"denom1:denom2", "denom3:denom4", "denom5:denom6"}
	for _, target := range voteTargets {
		keepers.OracleKeeper.WhitelistedPairs.Insert(ctx, target)
	}

	res, err := querier.VoteTargets(ctx, &types.QueryVoteTargetsRequest{})
	require.NoError(t, err)
	require.Equal(t, voteTargets, res.VoteTargets)
}
