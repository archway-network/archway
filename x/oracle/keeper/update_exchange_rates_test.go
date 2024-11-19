package keeper_test

import (
	"fmt"
	"math"
	"sort"
	"testing"

	"github.com/cometbft/cometbft/libs/rand"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/oracle/asset"
	"github.com/archway-network/archway/x/oracle/denoms"
	"github.com/archway-network/archway/x/oracle/keeper"
	"github.com/archway-network/archway/x/oracle/types"
)

func TestOracleThreshold(t *testing.T) {
	exchangeRates := types.ExchangeRateTuples{
		{
			Pair:         asset.Registry.Pair(denoms.BTC, denoms.USD),
			ExchangeRate: testExchangeRate,
		},
	}
	exchangeRateStr, err := exchangeRates.ToString()
	require.NoError(t, err)

	chain := e2eTesting.NewTestChain(t, 1,
		e2eTesting.WithValidatorsNum(5),
		e2eTesting.WithBondAmount(testStakingAmt.String()),
	)
	keepers := chain.GetApp().Keepers
	msgServer := keeper.NewMsgServerImpl(keepers.OracleKeeper)
	ctx := chain.GetContext()

	vals := chain.GetCurrentValSet().Validators
	AccAddrs := make([]sdk.AccAddress, len(vals))
	ValAddrs := make([]sdk.ValAddress, len(vals))
	for i := range vals {
		AccAddrs[i] = sdk.AccAddress(vals[i].Address)
		ValAddrs[i] = sdk.ValAddress(vals[i].Address)
	}

	params, err := keepers.OracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	params.VotePeriod = 1
	params.ExpirationBlocks = 0
	require.NoError(t, keepers.OracleKeeper.Params.Set(ctx, params))

	// Case 1.
	// Less than the threshold signs, exchange rate consensus fails
	for i := 0; i < 1; i++ {
		salt := fmt.Sprintf("%d", i)
		hash := types.GetAggregateVoteHash(salt, exchangeRateStr, ValAddrs[i])
		prevoteMsg := types.NewMsgAggregateExchangeRatePrevote(hash, AccAddrs[i], ValAddrs[i])
		voteMsg := types.NewMsgAggregateExchangeRateVote(salt, exchangeRateStr, AccAddrs[i], ValAddrs[i])

		_, err1 := msgServer.AggregateExchangeRatePrevote(ctx.WithBlockHeight(0), prevoteMsg)
		_, err2 := msgServer.AggregateExchangeRateVote(ctx.WithBlockHeight(1), voteMsg)
		require.NoError(t, err1)
		require.NoError(t, err2)
	}
	keepers.OracleKeeper.UpdateExchangeRates(ctx)
	_, err = keepers.OracleKeeper.ExchangeRates.Get(ctx.WithBlockHeight(1), exchangeRates[0].Pair)
	assert.Error(t, err)

	// Case 2.
	// More than the threshold signs, exchange rate consensus succeeds
	for i := 0; i < 4; i++ {
		salt := fmt.Sprintf("%d", i)
		hash := types.GetAggregateVoteHash(salt, exchangeRateStr, ValAddrs[i])
		prevoteMsg := types.NewMsgAggregateExchangeRatePrevote(hash, AccAddrs[i], ValAddrs[i])
		voteMsg := types.NewMsgAggregateExchangeRateVote(salt, exchangeRateStr, AccAddrs[i], ValAddrs[i])

		_, err1 := msgServer.AggregateExchangeRatePrevote(ctx.WithBlockHeight(0), prevoteMsg)
		_, err2 := msgServer.AggregateExchangeRateVote(ctx.WithBlockHeight(1), voteMsg)
		require.NoError(t, err1)
		require.NoError(t, err2)
	}
	keepers.OracleKeeper.UpdateExchangeRates(ctx)
	rate, err := keepers.OracleKeeper.ExchangeRates.Get(ctx, exchangeRates[0].Pair)
	require.NoError(t, err)
	assert.Equal(t, testExchangeRate, rate.ExchangeRate)

	// Case 3.
	// Increase voting power of absent validator, exchange rate consensus fails
	delegateAmount := testStakingAmt.MulRaw(8)
	delegateCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, delegateAmount))
	// topup not-bonded pool to withdraw for delegate
	err = keepers.BankKeeper.MintCoins(
		ctx,
		minttypes.ModuleName,
		delegateCoins,
	)
	require.NoError(t, err)
	err = keepers.BankKeeper.SendCoinsFromModuleToModule(
		ctx,
		minttypes.ModuleName,
		stakingtypes.NotBondedPoolName,
		delegateCoins,
	)
	require.NoError(t, err)
	val, err := keepers.StakingKeeper.GetValidator(ctx, ValAddrs[4])
	require.NoError(t, err)
	_, err = keepers.StakingKeeper.Delegate(ctx.WithBlockHeight(0), AccAddrs[4], delegateAmount, stakingtypes.Unbonded, val, false)
	require.NoError(t, err)

	for i := 0; i < 4; i++ {
		salt := fmt.Sprintf("%d", i)
		hash := types.GetAggregateVoteHash(salt, exchangeRateStr, ValAddrs[i])
		prevoteMsg := types.NewMsgAggregateExchangeRatePrevote(hash, AccAddrs[i], ValAddrs[i])
		voteMsg := types.NewMsgAggregateExchangeRateVote(salt, exchangeRateStr, AccAddrs[i], ValAddrs[i])

		_, err = msgServer.AggregateExchangeRatePrevote(ctx.WithBlockHeight(0), prevoteMsg)
		require.NoError(t, err)
		_, err = msgServer.AggregateExchangeRateVote(ctx.WithBlockHeight(1), voteMsg)
		require.NoError(t, err)
	}
	keepers.OracleKeeper.UpdateExchangeRates(ctx)
	_, err = keepers.OracleKeeper.ExchangeRates.Get(ctx, exchangeRates[0].Pair)
	assert.Error(t, err)
}

func TestResetExchangeRates(t *testing.T) {
	pair := asset.Registry.Pair(denoms.BTC, denoms.USD)
	chain := e2eTesting.NewTestChain(t, 1)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()

	emptyVotes := map[asset.Pair]types.ExchangeRateVotes{}
	validVotes := map[asset.Pair]types.ExchangeRateVotes{pair: {}}

	// Set expiration blocks to 10
	params, err := keepers.OracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	params.ExpirationBlocks = 10
	require.NoError(t, keepers.OracleKeeper.Params.Set(ctx, params))

	// Post a price at block 1
	keepers.OracleKeeper.SetPrice(ctx.WithBlockHeight(1), pair, testExchangeRate)

	// reset exchange rates at block 2
	// Price should still be there because not expired yet
	keepers.OracleKeeper.ClearExchangeRates(ctx.WithBlockHeight(2), emptyVotes)
	_, err = keepers.OracleKeeper.ExchangeRates.Get(ctx, pair)
	assert.NoError(t, err)

	// reset exchange rates at block 3 but pair is in votes
	// Price should be removed there because there was a valid votes
	keepers.OracleKeeper.ClearExchangeRates(ctx.WithBlockHeight(3), validVotes)
	_, err = keepers.OracleKeeper.ExchangeRates.Get(ctx, pair)
	assert.Error(t, err)

	// Post a price at block 69
	// reset exchange rates at block 79
	// Price should not be there anymore because expired
	keepers.OracleKeeper.SetPrice(ctx.WithBlockHeight(69), pair, testExchangeRate)
	keepers.OracleKeeper.ClearExchangeRates(ctx.WithBlockHeight(79), emptyVotes)

	_, err = keepers.OracleKeeper.ExchangeRates.Get(ctx, pair)
	assert.Error(t, err)
}

func TestOracleTally(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()

	votes := types.ExchangeRateVotes{}
	rates, valAddrs, stakingKeeper := types.GenerateRandomTestCase()
	keepers.OracleKeeper.StakingKeeper = stakingKeeper
	h := keeper.NewMsgServerImpl(keepers.OracleKeeper)

	params, err := keepers.OracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	params.VotePeriod = 1
	require.NoError(t, keepers.OracleKeeper.Params.Set(ctx, params))

	for i, rate := range rates {
		decExchangeRate := sdkmath.LegacyNewDecWithPrec(int64(rate*math.Pow10(OracleDecPrecision)), int64(OracleDecPrecision))
		exchangeRateStr, err := types.ExchangeRateTuples{
			{ExchangeRate: decExchangeRate, Pair: asset.Registry.Pair(denoms.BTC, denoms.USD)},
		}.ToString()
		require.NoError(t, err)

		salt := fmt.Sprintf("%d", i)
		hash := types.GetAggregateVoteHash(salt, exchangeRateStr, valAddrs[i])
		prevoteMsg := types.NewMsgAggregateExchangeRatePrevote(hash, sdk.AccAddress(valAddrs[i]), valAddrs[i])
		voteMsg := types.NewMsgAggregateExchangeRateVote(salt, exchangeRateStr, sdk.AccAddress(valAddrs[i]), valAddrs[i])

		_, err1 := h.AggregateExchangeRatePrevote(ctx.WithBlockHeight(0), prevoteMsg)
		_, err2 := h.AggregateExchangeRateVote(ctx.WithBlockHeight(1), voteMsg)
		require.NoError(t, err1)
		require.NoError(t, err2)

		power := testStakingAmt.QuoRaw(int64(6)).Int64()
		if decExchangeRate.IsZero() {
			power = int64(0)
		}

		vote := types.NewExchangeRateVote(
			decExchangeRate, asset.Registry.Pair(denoms.BTC, denoms.USD), valAddrs[i], power)
		votes = append(votes, vote)

		// change power of every three validator
		if i%3 == 0 {
			stakingKeeper.Validators()[i].SetConsensusPower(int64(i + 1))
		}
	}

	validatorPerformances := make(types.ValidatorPerformances)
	for _, valAddr := range valAddrs {
		val, err := stakingKeeper.Validator(ctx, valAddr)
		require.NoError(t, err)
		validatorPerformances[valAddr.String()] = types.NewValidatorPerformance(
			val.GetConsensusPower(sdk.DefaultPowerReduction),
			valAddr,
		)
	}
	sort.Sort(votes)
	weightedMedian := votes.WeightedMedianWithAssertion()
	standardDeviation := votes.StandardDeviation(weightedMedian)
	maxSpread := weightedMedian.Mul(keepers.OracleKeeper.RewardBand(ctx).QuoInt64(2))

	if standardDeviation.GT(maxSpread) {
		maxSpread = standardDeviation
	}

	expectedValidatorPerformances := make(types.ValidatorPerformances)
	for _, valAddr := range valAddrs {
		val, err := stakingKeeper.Validator(ctx, valAddr)
		require.NoError(t, err)
		expectedValidatorPerformances[valAddr.String()] = types.NewValidatorPerformance(
			val.GetConsensusPower(sdk.DefaultPowerReduction),
			valAddr,
		)
	}

	for _, vote := range votes {
		key := vote.Voter.String()
		validatorPerformance := expectedValidatorPerformances[key]
		if vote.ExchangeRate.GTE(weightedMedian.Sub(maxSpread)) &&
			vote.ExchangeRate.LTE(weightedMedian.Add(maxSpread)) {
			validatorPerformance.RewardWeight += vote.Power
			validatorPerformance.WinCount++
		} else if !vote.ExchangeRate.IsPositive() {
			validatorPerformance.AbstainCount++
		} else {
			validatorPerformance.MissCount++
		}
		expectedValidatorPerformances[key] = validatorPerformance
	}

	tallyMedian := keeper.Tally(
		votes, keepers.OracleKeeper.RewardBand(ctx), validatorPerformances,
	)

	assert.Equal(t, expectedValidatorPerformances, validatorPerformances)
	assert.Equal(t, tallyMedian.MulInt64(100).TruncateInt(), weightedMedian.MulInt64(100).TruncateInt())
	assert.NotEqualValues(t, 0, validatorPerformances.TotalRewardWeight(), validatorPerformances.String())
}

func TestOracleRewardBand(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1, e2eTesting.WithValidatorsNum(4))
	keepers := chain.GetApp().Keepers
	msgServer := keeper.NewMsgServerImpl(keepers.OracleKeeper)
	ctx := chain.GetContext()

	vals := chain.GetCurrentValSet().Validators
	ValAddrs := make([]sdk.ValAddress, len(vals))
	for i := range vals {
		ValAddrs[i] = sdk.ValAddress(vals[i].Address)
	}

	params, err := keepers.OracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	params.VotePeriod = 1
	params.Whitelist = []asset.Pair{asset.Registry.Pair(denoms.ATOM, denoms.USD)}
	require.NoError(t, keepers.OracleKeeper.Params.Set(ctx, params))

	// clear pairs to reset vote targets
	err = keepers.OracleKeeper.WhitelistedPairs.Clear(ctx, nil)
	require.NoError(t, err)
	err = keepers.OracleKeeper.WhitelistedPairs.Set(ctx, asset.Registry.Pair(denoms.ATOM, denoms.USD))
	require.NoError(t, err)

	rewardSpread := testExchangeRate.Mul(keepers.OracleKeeper.RewardBand(ctx).QuoInt64(2))

	// Account 1, atom:usd
	MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
		{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: testExchangeRate.Sub(rewardSpread)},
	}, vals[0])

	// Account 2, atom:usd
	MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
		{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: testExchangeRate},
	}, vals[1])

	// Account 3, atom:usd
	MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
		{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: testExchangeRate},
	}, vals[2])

	// Account 4, atom:usd
	MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
		{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: testExchangeRate.Add(rewardSpread)},
	}, vals[3])

	keepers.OracleKeeper.UpdateExchangeRates(ctx)

	counter, err := keepers.OracleKeeper.MissCounters.Get(ctx, ValAddrs[0])
	if err != nil {
		counter = 0
	}
	assert.Equal(t, uint64(0), counter)
	counter, err = keepers.OracleKeeper.MissCounters.Get(ctx, ValAddrs[1])
	if err != nil {
		counter = 0
	}
	assert.Equal(t, uint64(0), counter)
	counter, err = keepers.OracleKeeper.MissCounters.Get(ctx, ValAddrs[2])
	if err != nil {
		counter = 0
	}
	assert.Equal(t, uint64(0), counter)
	counter, err = keepers.OracleKeeper.MissCounters.Get(ctx, ValAddrs[3])
	if err != nil {
		counter = 0
	}
	assert.Equal(t, uint64(0), counter)

	// Account 1 will miss the vote due to raward band condition
	// Account 1, atom:usd
	MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
		{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: testExchangeRate.Sub(rewardSpread.Add(sdkmath.LegacyOneDec()))},
	}, vals[0])

	// Account 2, atom:usd
	MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
		{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: testExchangeRate},
	}, vals[1])

	// Account 3, atom:usd
	MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
		{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: testExchangeRate},
	}, vals[2])

	// Account 4, atom:usd
	MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
		{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: testExchangeRate.Add(rewardSpread)},
	}, vals[3])

	keepers.OracleKeeper.UpdateExchangeRates(ctx)

	counter, err = keepers.OracleKeeper.MissCounters.Get(ctx, ValAddrs[0])
	if err != nil {
		counter = 0
	}
	assert.Equal(t, uint64(1), counter)
	counter, err = keepers.OracleKeeper.MissCounters.Get(ctx, ValAddrs[1])
	if err != nil {
		counter = 0
	}
	assert.Equal(t, uint64(0), counter)
	counter, err = keepers.OracleKeeper.MissCounters.Get(ctx, ValAddrs[2])
	if err != nil {
		counter = 0
	}
	assert.Equal(t, uint64(0), counter)
	counter, err = keepers.OracleKeeper.MissCounters.Get(ctx, ValAddrs[3])
	if err != nil {
		counter = 0
	}
	assert.Equal(t, uint64(0), counter)
}

/* TODO(Mercilex): not appliable right now: https://github.com/archway-network/archway/issues/805
func TestOracleMultiRewardDistribution(t *testing.T) {
	input, h := setup(t)

	// SDR and KRW have the same voting power, but KRW has been chosen as referencepair by alphabetical order.
	// Account 1, SDR, KRW
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.Pairbtc:usd.String(), ExchangeRate: randomExchangeRate}, {Pair: common.Pairatom:usd.String(), ExchangeRate: randomExchangeRate}}, vals[0])

	// Account 2, SDR
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.Pairbtc:usd.String(), ExchangeRate: randomExchangeRate}}, vals[1])

	// Account 3, KRW
	makeAggregatePrevoteAndVote(t, input, h, 0, types.ExchangeRateTuples{{Pair: common.Pairbtc:usd.String(), ExchangeRate: randomExchangeRate}}, vals[2])

	rewardAmt := math.NewInt(1e6)
	err := input.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(denoms.Gov, rewardAmt)))
	require.NoError(t, err)

	keepers.OracleKeeper.UpdateExchangeRates(ctx)

	rewardDistributedWindow := keepers.OracleKeeper.RewardDistributionWindow(ctx)

	expectedRewardAmt := math.LegacyNewDecFromInt(rewardAmt.QuoRaw(3).MulRaw(2)).QuoInt64(int64(rewardDistributedWindow)).TruncateInt()
	expectedRewardAmt2 := math.ZeroInt() // even vote power is same KRW with SDR, KRW chosen referenceTerra because alphabetical order
	expectedRewardAmt3 := math.LegacyNewDecFromInt(rewardAmt.QuoRaw(3)).QuoInt64(int64(rewardDistributedWindow)).TruncateInt()

	rewards := keepers.DistrKeeper.GetValidatorOutstandingRewards(ctx.WithBlockHeight(2), ValAddrs[0])
	assert.Equal(t, expectedRewardAmt, rewards.Rewards.AmountOf(denoms.Gov).TruncateInt())
	rewards = keepers.DistrKeeper.GetValidatorOutstandingRewards(ctx.WithBlockHeight(2), ValAddrs[1])
	assert.Equal(t, expectedRewardAmt2, rewards.Rewards.AmountOf(denoms.Gov).TruncateInt())
	rewards = keepers.DistrKeeper.GetValidatorOutstandingRewards(ctx.WithBlockHeight(2), ValAddrs[2])
	assert.Equal(t, expectedRewardAmt3, rewards.Rewards.AmountOf(denoms.Gov).TruncateInt())
}
*/

func TestOracleExchangeRate(t *testing.T) {
	// The following scenario tests four validators providing prices for eth:usd, atom:usd, and btc:usd.
	// eth:usd and atom:usd pass, but btc:usd fails due to not enough validators voting.
	chain := e2eTesting.NewTestChain(t, 1, e2eTesting.WithValidatorsNum(4))
	keepers := chain.GetApp().Keepers
	msgServer := keeper.NewMsgServerImpl(keepers.OracleKeeper)
	ctx := chain.GetContext()

	vals := chain.GetCurrentValSet().Validators
	ValAddrs := make([]sdk.ValAddress, len(vals))
	for i := range vals {
		ValAddrs[i] = sdk.ValAddress(vals[i].Address)
	}

	params, err := keepers.OracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	params.VotePeriod = 1
	require.NoError(t, keepers.OracleKeeper.Params.Set(ctx, params))

	atomUsdExchangeRate := sdkmath.LegacyNewDec(1000000)
	ethUsdExchangeRate := sdkmath.LegacyNewDec(1000000)
	btcusdExchangeRate := sdkmath.LegacyNewDec(1e6)

	// Account 1, eth:usd, atom:usd, btc:usd
	MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
		{Pair: asset.Registry.Pair(denoms.ETH, denoms.USD), ExchangeRate: ethUsdExchangeRate},
		{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: atomUsdExchangeRate},
		{Pair: asset.Registry.Pair(denoms.BTC, denoms.USD), ExchangeRate: btcusdExchangeRate},
	}, vals[0])

	// Account 2, eth:usd, atom:usd
	MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
		{Pair: asset.Registry.Pair(denoms.ETH, denoms.USD), ExchangeRate: ethUsdExchangeRate},
		{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: atomUsdExchangeRate},
	}, vals[1])

	// Account 3, eth:usd, atom:usd, btc:usd(abstain)
	MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
		{Pair: asset.Registry.Pair(denoms.ETH, denoms.USD), ExchangeRate: ethUsdExchangeRate},
		{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: atomUsdExchangeRate},
		{Pair: asset.Registry.Pair(denoms.BTC, denoms.USD), ExchangeRate: sdkmath.LegacyZeroDec()},
	}, vals[2])

	// Account 4, eth:usd, atom:usd, btc:usd
	MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
		{Pair: asset.Registry.Pair(denoms.ETH, denoms.USD), ExchangeRate: ethUsdExchangeRate},
		{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: atomUsdExchangeRate},
		{Pair: asset.Registry.Pair(denoms.BTC, denoms.USD), ExchangeRate: sdkmath.LegacyZeroDec()},
	}, vals[3])

	ethUsdRewards := sdk.NewInt64Coin("ETHREWARD", 2000000)
	atomUsdRewards := sdk.NewInt64Coin("ATOMREWARD", 3000000)

	require.NoError(t, keepers.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(ethUsdRewards)))
	require.NoError(t, keepers.OracleKeeper.AllocateRewards(ctx, minttypes.ModuleName, sdk.NewCoins(ethUsdRewards), 1))
	require.NoError(t, keepers.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(atomUsdRewards)))
	require.NoError(t, keepers.OracleKeeper.AllocateRewards(ctx, minttypes.ModuleName, sdk.NewCoins(atomUsdRewards), 1))

	keepers.OracleKeeper.UpdateExchangeRates(ctx)

	// total reward pool for the current vote period is 1* common.TO_MICRO for eth:usd and 1* common.TO_MICRO for atom:usd
	// val 1,2,3,4 all won on 2 pairs
	// so total votes are 2 * 2 + 2 + 2 = 8
	expectedRewardAmt := sdk.NewDecCoinsFromCoins(ethUsdRewards, atomUsdRewards).
		QuoDec(sdkmath.LegacyNewDec(8)). // total votes
		MulDec(sdkmath.LegacyNewDec(2))  // votes won by val1 and val2
	rewards, err := keepers.DistrKeeper.GetValidatorOutstandingRewards(ctx.WithBlockHeight(2), ValAddrs[0])
	require.NoError(t, err)
	assert.Equalf(t, expectedRewardAmt, rewards.Rewards, "%s <-> %s", expectedRewardAmt, rewards.Rewards)
	rewards, err = keepers.DistrKeeper.GetValidatorOutstandingRewards(ctx.WithBlockHeight(2), ValAddrs[1])
	require.NoError(t, err)
	assert.Equalf(t, expectedRewardAmt, rewards.Rewards, "%s <-> %s", expectedRewardAmt, rewards.Rewards)
	rewards, err = keepers.DistrKeeper.GetValidatorOutstandingRewards(ctx.WithBlockHeight(2), ValAddrs[2])
	require.NoError(t, err)
	assert.Equalf(t, expectedRewardAmt, rewards.Rewards, "%s <-> %s", expectedRewardAmt, rewards.Rewards)
	rewards, err = keepers.DistrKeeper.GetValidatorOutstandingRewards(ctx.WithBlockHeight(2), ValAddrs[3])
	require.NoError(t, err)
	assert.Equalf(t, expectedRewardAmt, rewards.Rewards, "%s <-> %s", expectedRewardAmt, rewards.Rewards)
}

func TestOracleRandomPrices(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1, e2eTesting.WithValidatorsNum(4))
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()
	msgServer := keeper.NewMsgServerImpl(keepers.OracleKeeper)

	vals := chain.GetCurrentValSet().Validators

	params, err := keepers.OracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	params.VotePeriod = 1
	require.NoError(t, keepers.OracleKeeper.Params.Set(ctx, params))

	for i := 0; i < 100; i++ {
		for _, val := range vals {
			MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
				{Pair: asset.Registry.Pair(denoms.ETH, denoms.USD), ExchangeRate: sdkmath.LegacyNewDec(int64(rand.Uint64() % 1e6))},
				{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: sdkmath.LegacyNewDec(int64(rand.Uint64() % 1e6))},
			}, val)
		}

		require.NotPanics(t, func() {
			keepers.OracleKeeper.UpdateExchangeRates(ctx)
		})
	}
}

func TestWhitelistedPairs(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1, e2eTesting.WithValidatorsNum(5))
	keepers := chain.GetApp().Keepers
	msgServer := keeper.NewMsgServerImpl(keepers.OracleKeeper)
	ctx := chain.GetContext()

	vals := chain.GetCurrentValSet().Validators
	ValAddrs := make([]sdk.ValAddress, len(vals))
	for i := range vals {
		ValAddrs[i] = sdk.ValAddress(vals[i].Address)
	}

	params, err := keepers.OracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	params.VotePeriod = 1
	require.NoError(t, keepers.OracleKeeper.Params.Set(ctx, params))

	t.Log("whitelist ONLY atom:usd")
	err = keepers.OracleKeeper.WhitelistedPairs.Clear(ctx, nil)
	require.NoError(t, err)
	err = keepers.OracleKeeper.WhitelistedPairs.Set(ctx, asset.Registry.Pair(denoms.ATOM, denoms.USD))
	require.NoError(t, err)

	t.Log("vote and prevote from all vals on atom:usd")
	priceVoteFromVal := func(valIdx int, block int64) {
		MakeAggregatePrevoteAndVote(t, ctx, msgServer, block, types.ExchangeRateTuples{{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: testExchangeRate}}, vals[valIdx])
	}
	block := int64(0)
	priceVoteFromVal(0, block)
	priceVoteFromVal(1, block)
	priceVoteFromVal(2, block)
	priceVoteFromVal(3, block)

	t.Log("whitelist btc:usd for next vote period")
	params.Whitelist = []asset.Pair{asset.Registry.Pair(denoms.ATOM, denoms.USD), asset.Registry.Pair(denoms.BTC, denoms.USD)}
	require.NoError(t, keepers.OracleKeeper.Params.Set(ctx, params))
	keepers.OracleKeeper.UpdateExchangeRates(ctx)

	t.Log("assert: no miss counts for all vals")
	counter, err := keepers.OracleKeeper.MissCounters.Get(ctx, ValAddrs[0])
	if err != nil {
		counter = 0
	}
	assert.Equal(t, uint64(0), counter)
	counter, err = keepers.OracleKeeper.MissCounters.Get(ctx, ValAddrs[1])
	if err != nil {
		counter = 0
	}
	assert.Equal(t, uint64(0), counter)
	counter, err = keepers.OracleKeeper.MissCounters.Get(ctx, ValAddrs[2])
	if err != nil {
		counter = 0
	}
	assert.Equal(t, uint64(0), counter)
	counter, err = keepers.OracleKeeper.MissCounters.Get(ctx, ValAddrs[3])
	if err != nil {
		counter = 0
	}
	assert.Equal(t, uint64(0), counter)

	t.Log("whitelisted pairs are {atom:usd, btc:usd}")
	pairs, err := keepers.OracleKeeper.GetWhitelistedPairs(ctx)
	require.NoError(t, err)
	assert.Equal(t,
		[]asset.Pair{
			asset.Registry.Pair(denoms.ATOM, denoms.USD),
			asset.Registry.Pair(denoms.BTC, denoms.USD),
		},
		pairs,
	)

	t.Log("vote from vals 0-3 on atom:usd (but not btc:usd)")
	priceVoteFromVal(0, block)
	priceVoteFromVal(1, block)
	priceVoteFromVal(2, block)
	priceVoteFromVal(3, block)

	t.Log("delete btc:usd for next vote period")
	params.Whitelist = []asset.Pair{asset.Registry.Pair(denoms.ATOM, denoms.USD)}
	require.NoError(t, keepers.OracleKeeper.Params.Set(ctx, params))
	perfs := keepers.OracleKeeper.UpdateExchangeRates(ctx)

	t.Log("validators 0-3 all voted -> expect win")
	for valIdx := 0; valIdx < 4; valIdx++ {
		perf := perfs[ValAddrs[valIdx].String()]
		assert.EqualValues(t, 1, perf.WinCount)
		assert.EqualValues(t, 1, perf.AbstainCount)
		assert.EqualValues(t, 0, perf.MissCount)
	}
	t.Log("validators 4 didn't vote -> expect abstain")
	perf := perfs[ValAddrs[4].String()]
	assert.EqualValues(t, 0, perf.WinCount)
	assert.EqualValues(t, 2, perf.AbstainCount)
	assert.EqualValues(t, 0, perf.MissCount)

	t.Log("btc:usd must be deleted")
	pairs, err = keepers.OracleKeeper.GetWhitelistedPairs(ctx)
	require.NoError(t, err)
	assert.Equal(t,
		[]asset.Pair{asset.Registry.Pair(denoms.ATOM, denoms.USD)},
		pairs,
	)
	has, err := keepers.OracleKeeper.WhitelistedPairs.Has(ctx, asset.Registry.Pair(denoms.BTC, denoms.USD))
	require.NoError(t, err)
	require.False(t, has)

	t.Log("vote from vals 0-3 on atom:usd")
	priceVoteFromVal(0, block)
	priceVoteFromVal(1, block)
	priceVoteFromVal(2, block)
	priceVoteFromVal(3, block)
	perfs = keepers.OracleKeeper.UpdateExchangeRates(ctx)

	t.Log("Although validators 0-2 voted, it's for the same period -> expect abstains for everyone")
	for valIdx := 0; valIdx < 4; valIdx++ {
		perf := perfs[ValAddrs[valIdx].String()]
		assert.EqualValues(t, 1, perf.WinCount)
		assert.EqualValues(t, 0, perf.AbstainCount)
		assert.EqualValues(t, 0, perf.MissCount)
	}
	perf = perfs[ValAddrs[4].String()]
	assert.EqualValues(t, 0, perf.WinCount)
	assert.EqualValues(t, 1, perf.AbstainCount)
	assert.EqualValues(t, 0, perf.MissCount)
}
