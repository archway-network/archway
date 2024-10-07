package keeper_test

import (
	"sort"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/types/set"
	"github.com/archway-network/archway/x/oracle/asset"
	"github.com/archway-network/archway/x/oracle/denoms"
	"github.com/archway-network/archway/x/oracle/keeper"
	"github.com/archway-network/archway/x/oracle/types"
)

func TestGroupVotesByPair(t *testing.T) {
	power := int64(100)

	chain := e2eTesting.NewTestChain(t, 1,
		e2eTesting.WithValidatorsNum(3),
	)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()

	vals := chain.GetCurrentValSet().Validators
	ValAddrs := make([]sdk.ValAddress, len(vals))
	for i := range vals {
		ValAddrs[i] = sdk.ValAddress(vals[i].Address)
	}

	keepers.StakingKeeper.EndBlocker(ctx)

	pairBtc := asset.Registry.Pair(denoms.BTC, denoms.NUSD)
	pairEth := asset.Registry.Pair(denoms.ETH, denoms.NUSD)
	btcVotes := types.ExchangeRateVotes{
		{Pair: pairBtc, ExchangeRate: math.LegacyNewDec(17), Voter: ValAddrs[0], Power: power},
		{Pair: pairBtc, ExchangeRate: math.LegacyNewDec(10), Voter: ValAddrs[1], Power: power},
		{Pair: pairBtc, ExchangeRate: math.LegacyNewDec(6), Voter: ValAddrs[2], Power: power},
	}
	ethVotes := types.ExchangeRateVotes{
		{Pair: pairEth, ExchangeRate: math.LegacyNewDec(1_000), Voter: ValAddrs[0], Power: power},
		{Pair: pairEth, ExchangeRate: math.LegacyNewDec(1_300), Voter: ValAddrs[1], Power: power},
		{Pair: pairEth, ExchangeRate: math.LegacyNewDec(2_000), Voter: ValAddrs[2], Power: power},
	}

	for i, v := range btcVotes {
		keepers.OracleKeeper.Votes.Set(
			ctx,
			ValAddrs[i],
			types.NewAggregateExchangeRateVote(
				types.ExchangeRateTuples{
					{Pair: v.Pair, ExchangeRate: v.ExchangeRate},
					{Pair: ethVotes[i].Pair, ExchangeRate: ethVotes[i].ExchangeRate},
				},
				ValAddrs[i],
			),
		)
	}

	// organize votes by pair
	pairVotes := keepers.OracleKeeper.GroupVotesByPair(ctx, types.ValidatorPerformances{
		ValAddrs[0].String(): {
			Power:      power,
			WinCount:   0,
			ValAddress: ValAddrs[0],
		},
		ValAddrs[1].String(): {
			Power:      power,
			WinCount:   0,
			ValAddress: ValAddrs[1],
		},
		ValAddrs[2].String(): {
			Power:      power,
			WinCount:   0,
			ValAddress: ValAddrs[2],
		},
	})

	// sort each votes for comparison
	sort.Sort(btcVotes)
	sort.Sort(ethVotes)
	sort.Sort(pairVotes[asset.Registry.Pair(denoms.BTC, denoms.NUSD)])
	sort.Sort(pairVotes[asset.Registry.Pair(denoms.ETH, denoms.NUSD)])

	require.Equal(t, btcVotes, pairVotes[asset.Registry.Pair(denoms.BTC, denoms.NUSD)])
	require.Equal(t, ethVotes, pairVotes[asset.Registry.Pair(denoms.ETH, denoms.NUSD)])
}

func TestClearVotesAndPrevotes(t *testing.T) {
	power := int64(100)

	chain := e2eTesting.NewTestChain(t, 1, e2eTesting.WithValidatorsNum(3))
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()

	vals := chain.GetCurrentValSet().Validators
	ValAddrs := make([]sdk.ValAddress, len(vals))
	for i := range vals {
		ValAddrs[i] = sdk.ValAddress(vals[i].Address)
	}

	keepers.StakingKeeper.EndBlocker(ctx)

	btcVotes := types.ExchangeRateVotes{
		types.NewExchangeRateVote(math.LegacyNewDec(17), asset.Registry.Pair(denoms.BTC, denoms.NUSD), ValAddrs[0], power),
		types.NewExchangeRateVote(math.LegacyNewDec(10), asset.Registry.Pair(denoms.BTC, denoms.NUSD), ValAddrs[1], power),
		types.NewExchangeRateVote(math.LegacyNewDec(6), asset.Registry.Pair(denoms.BTC, denoms.NUSD), ValAddrs[2], power),
	}
	ethVotes := types.ExchangeRateVotes{
		types.NewExchangeRateVote(math.LegacyNewDec(1000), asset.Registry.Pair(denoms.ETH, denoms.NUSD), ValAddrs[0], power),
		types.NewExchangeRateVote(math.LegacyNewDec(1300), asset.Registry.Pair(denoms.ETH, denoms.NUSD), ValAddrs[1], power),
		types.NewExchangeRateVote(math.LegacyNewDec(2000), asset.Registry.Pair(denoms.ETH, denoms.NUSD), ValAddrs[2], power),
	}

	for i := range btcVotes {
		keepers.OracleKeeper.Prevotes.Set(ctx, ValAddrs[i], types.AggregateExchangeRatePrevote{
			Hash:        "",
			Voter:       ValAddrs[i].String(),
			SubmitBlock: uint64(ctx.BlockHeight()),
		})

		keepers.OracleKeeper.Votes.Set(ctx, ValAddrs[i],
			types.NewAggregateExchangeRateVote(types.ExchangeRateTuples{
				{Pair: btcVotes[i].Pair, ExchangeRate: btcVotes[i].ExchangeRate},
				{Pair: ethVotes[i].Pair, ExchangeRate: ethVotes[i].ExchangeRate},
			}, ValAddrs[i]))
	}

	keepers.OracleKeeper.ClearVotesAndPrevotes(ctx, 10)

	prevoteCounterIter, err := keepers.OracleKeeper.Prevotes.Iterate(ctx, nil)
	require.NoError(t, err)
	prevoteCounterKeys, err := prevoteCounterIter.Keys()
	require.NoError(t, err)
	prevoteCounter := len(prevoteCounterKeys)

	voteCounterIter, err := keepers.OracleKeeper.Votes.Iterate(ctx, nil)
	require.NoError(t, err)
	voteCounterKeys, err := voteCounterIter.Keys()
	require.NoError(t, err)
	voteCounter := len(voteCounterKeys)

	require.Equal(t, prevoteCounter, 3)
	require.Equal(t, voteCounter, 0)

	// vote period starts at b=10, clear the votes at b=0 and below.
	keepers.OracleKeeper.ClearVotesAndPrevotes(ctx.WithBlockHeight(ctx.BlockHeight()+10), 10)
	prevoteCounterIter, err = keepers.OracleKeeper.Prevotes.Iterate(ctx, nil)
	require.NoError(t, err)
	prevoteCounterKeys, err = prevoteCounterIter.Keys()
	require.NoError(t, err)
	prevoteCounter = len(prevoteCounterKeys)
	require.Equal(t, prevoteCounter, 0)
}

func TestFuzzTally(t *testing.T) {
	validators := map[string]int64{}

	f := fuzz.New().NilChance(0).Funcs(
		func(e *math.LegacyDec, c fuzz.Continue) {
			*e = math.LegacyNewDec(c.Int63())
		},
		func(e *map[string]int64, c fuzz.Continue) {
			numValidators := c.Intn(100) + 5

			for i := 0; i < numValidators; i++ {
				(*e)[sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address()).String()] = c.Int63n(100)
			}
		},
		func(e *types.ValidatorPerformances, c fuzz.Continue) {
			for validator, power := range validators {
				addr, err := sdk.ValAddressFromBech32(validator)
				require.NoError(t, err)
				(*e)[validator] = types.NewValidatorPerformance(power, addr)
			}
		},
		func(e *types.ExchangeRateVotes, c fuzz.Continue) {
			votes := types.ExchangeRateVotes{}
			for addr, power := range validators {
				addr, _ := sdk.ValAddressFromBech32(addr)

				var rate math.LegacyDec
				c.Fuzz(&rate)

				votes = append(votes, types.NewExchangeRateVote(rate, asset.NewPair(c.RandString(), c.RandString()), addr, power))
			}

			*e = votes
		},
	)

	// set random pairs and validators
	f.Fuzz(&validators)

	claimMap := types.ValidatorPerformances{}
	f.Fuzz(&claimMap)

	votes := types.ExchangeRateVotes{}
	f.Fuzz(&votes)

	var rewardBand math.LegacyDec
	f.Fuzz(&rewardBand)

	require.NotPanics(t, func() {
		keeper.Tally(votes, rewardBand, claimMap)
	})
}

type VoteMap = map[asset.Pair]types.ExchangeRateVotes

func TestRemoveInvalidBallots(t *testing.T) {
	testCases := []struct {
		name    string
		voteMap VoteMap
	}{
		{
			name: "empty key, empty votes",
			voteMap: VoteMap{
				"": types.ExchangeRateVotes{},
			},
		},
		{
			name: "nonempty key, empty votes",
			voteMap: VoteMap{
				"xxx": types.ExchangeRateVotes{},
			},
		},
		{
			name: "nonempty keys, empty votes",
			voteMap: VoteMap{
				"xxx":    types.ExchangeRateVotes{},
				"abc123": types.ExchangeRateVotes{},
			},
		},
		{
			name: "mixed empty keys, empty votes",
			voteMap: VoteMap{
				"xxx":    types.ExchangeRateVotes{},
				"":       types.ExchangeRateVotes{},
				"abc123": types.ExchangeRateVotes{},
				"0x":     types.ExchangeRateVotes{},
			},
		},
		{
			name: "empty key, nonempty votes, not whitelisted",
			voteMap: VoteMap{
				"": types.ExchangeRateVotes{
					{Pair: "", ExchangeRate: math.LegacyZeroDec(), Voter: sdk.ValAddress{}, Power: 0},
				},
			},
		},
		{
			name: "nonempty key, nonempty votes, whitelisted",
			voteMap: VoteMap{
				"x": types.ExchangeRateVotes{
					{Pair: "x", ExchangeRate: math.LegacyZeroDec(), Voter: sdk.ValAddress{123}, Power: 5},
				},
				asset.Registry.Pair(denoms.BTC, denoms.NUSD): types.ExchangeRateVotes{
					{Pair: asset.Registry.Pair(denoms.BTC, denoms.NUSD), ExchangeRate: math.LegacyZeroDec(), Voter: sdk.ValAddress{123}, Power: 5},
				},
				asset.Registry.Pair(denoms.ETH, denoms.NUSD): types.ExchangeRateVotes{
					{Pair: asset.Registry.Pair(denoms.BTC, denoms.NUSD), ExchangeRate: math.LegacyZeroDec(), Voter: sdk.ValAddress{123}, Power: 5},
				},
			},
		},
	}

	for i, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			chain := e2eTesting.NewTestChain(t, i)
			keepers := chain.GetApp().Keepers
			assert.NotPanics(t, func() {
				keepers.OracleKeeper.RemoveInvalidVotes(chain.GetContext(), tc.voteMap, set.New[asset.Pair](
					asset.NewPair(denoms.BTC, denoms.NUSD),
					asset.NewPair(denoms.ETH, denoms.NUSD),
				))
			}, "voteMap: %v", tc.voteMap)
		})
	}
}

func TestFuzzPickReferencePair(t *testing.T) {
	var pairs []asset.Pair

	f := fuzz.New().NilChance(0).Funcs(
		func(e *asset.Pair, c fuzz.Continue) {
			*e = asset.NewPair(RandLetters(5), RandLetters(5))
		},
		func(e *[]asset.Pair, c fuzz.Continue) {
			numPairs := c.Intn(100) + 5

			for i := 0; i < numPairs; i++ {
				*e = append(*e, asset.NewPair(RandLetters(5), RandLetters(5)))
			}
		},
		func(e *math.LegacyDec, c fuzz.Continue) {
			*e = math.LegacyNewDec(c.Int63())
		},
		func(e *map[asset.Pair]math.LegacyDec, c fuzz.Continue) {
			for _, pair := range pairs {
				var rate math.LegacyDec
				c.Fuzz(&rate)

				(*e)[pair] = rate
			}
		},
		func(e *map[string]int64, c fuzz.Continue) {
			for i := 0; i < 5+c.Intn(100); i++ {
				(*e)[sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address()).String()] = int64(c.Intn(100) + 1)
			}
		},
		func(e *map[asset.Pair]types.ExchangeRateVotes, c fuzz.Continue) {
			validators := map[string]int64{}
			c.Fuzz(&validators)

			for _, pair := range pairs {
				votes := types.ExchangeRateVotes{}

				for addr, power := range validators {
					addr, _ := sdk.ValAddressFromBech32(addr)

					var rate math.LegacyDec
					c.Fuzz(&rate)

					votes = append(votes, types.NewExchangeRateVote(rate, pair, addr, power))
				}

				(*e)[pair] = votes
			}
		},
	)

	// set random pairs
	f.Fuzz(&pairs)

	chain := e2eTesting.NewTestChain(t, 1)

	// test OracleKeeper.Pairs.Insert
	voteTargets := set.Set[asset.Pair]{}
	f.Fuzz(&voteTargets)
	whitelistedPairs := make(set.Set[asset.Pair])

	for key := range voteTargets {
		whitelistedPairs.Add(key)
	}

	// test OracleKeeper.RemoveInvalidBallots
	voteMap := map[asset.Pair]types.ExchangeRateVotes{}
	f.Fuzz(&voteMap)

	assert.NotPanics(t, func() {
		chain.GetApp().Keepers.OracleKeeper.RemoveInvalidVotes(chain.GetContext(), voteMap, whitelistedPairs)
	}, "voteMap: %v", voteMap)
}

func TestZeroBallotPower(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1, e2eTesting.WithValidatorsNum(3))

	vals := chain.GetCurrentValSet().Validators
	ValAddrs := make([]sdk.ValAddress, len(vals))
	for i := range vals {
		ValAddrs[i] = sdk.ValAddress(vals[i].Address)
	}

	btcVotess := types.ExchangeRateVotes{
		types.NewExchangeRateVote(math.LegacyNewDec(17), asset.Registry.Pair(denoms.BTC, denoms.NUSD), ValAddrs[0], 0),
		types.NewExchangeRateVote(math.LegacyNewDec(10), asset.Registry.Pair(denoms.BTC, denoms.NUSD), ValAddrs[1], 0),
		types.NewExchangeRateVote(math.LegacyNewDec(6), asset.Registry.Pair(denoms.BTC, denoms.NUSD), ValAddrs[2], 0),
	}

	assert.False(t, keeper.IsPassingVoteThreshold(btcVotess, math.ZeroInt(), 0))
}
