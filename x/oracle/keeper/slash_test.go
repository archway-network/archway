package keeper_test

import (
	"testing"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	"github.com/NibiruChain/collections"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/common/asset"
	"github.com/archway-network/archway/x/common/denoms"
	"github.com/archway-network/archway/x/oracle/keeper"
	"github.com/archway-network/archway/x/oracle/types"
)

func TestSlashAndResetMissCounters(t *testing.T) {
	// initial setup
	accNum := 2
	amt := sdk.TokensFromConsensusPower(50, sdk.DefaultPowerReduction)
	InitTokens := sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)

	chain := e2eTesting.NewTestChain(t, 1,
		e2eTesting.WithValidatorsNum(accNum),
		e2eTesting.WithGenAccounts(accNum),
		e2eTesting.WithBondAmount(amt.String()),
		e2eTesting.WithGenDefaultCoinBalance(InitTokens.String()),
	)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()

	AccAddrs := make([]sdk.AccAddress, accNum)
	ValAddrs := make([]sdk.ValAddress, accNum)
	for i := 0; i < accNum; i++ {
		AccAddrs[i] = chain.GetAccount(i).Address
	}
	for i, val := range chain.GetCurrentValSet().Validators {
		ValAddrs[i] = sdk.ValAddress(val.Address)
	}

	params, err := keepers.OracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	params.VotePeriod = 1
	keepers.OracleKeeper.Params.Set(ctx, params)

	keepers.StakingKeeper.EndBlocker(ctx)

	stakingparams, err := keepers.StakingKeeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(
		t,
		sdk.NewCoins(sdk.NewCoin(stakingparams.BondDenom, InitTokens.Sub(amt))),
		keepers.BankKeeper.GetAllBalances(ctx, AccAddrs[0]),
	)
	validator, err := keepers.StakingKeeper.GetValidator(ctx, ValAddrs[0])
	require.NoError(t, err)
	require.Equal(t, amt, validator.GetBondedTokens())
	require.Equal(
		t,
		sdk.NewCoins(sdk.NewCoin(stakingparams.BondDenom, InitTokens.Sub(amt))),
		keepers.BankKeeper.GetAllBalances(ctx, AccAddrs[1]),
	)
	validator1, err := keepers.StakingKeeper.GetValidator(ctx, ValAddrs[1])
	require.NoError(t, err)
	require.Equal(t, amt, validator1.GetBondedTokens())

	votePeriodsPerWindow := math.LegacyNewDec(int64(keepers.OracleKeeper.SlashWindow(ctx))).QuoInt64(int64(keepers.OracleKeeper.VotePeriod(ctx))).TruncateInt64()
	slashFraction := keepers.OracleKeeper.SlashFraction(ctx)
	minValidVotes := keepers.OracleKeeper.MinValidPerWindow(ctx).MulInt64(votePeriodsPerWindow).Ceil().TruncateInt64()
	// Case 1, no slash
	keepers.OracleKeeper.MissCounters.Insert(ctx, ValAddrs[0], uint64(votePeriodsPerWindow-minValidVotes))
	keepers.OracleKeeper.SlashAndResetMissCounters(ctx)
	keepers.StakingKeeper.EndBlocker(ctx)

	validator, _ = keepers.StakingKeeper.GetValidator(ctx, ValAddrs[0])
	require.NoError(t, err)
	require.Equal(t, amt, validator.GetBondedTokens())

	// Case 2, slash
	keepers.OracleKeeper.MissCounters.Insert(ctx, ValAddrs[0], uint64(votePeriodsPerWindow-minValidVotes+1))
	keepers.OracleKeeper.SlashAndResetMissCounters(ctx)
	validator, _ = keepers.StakingKeeper.GetValidator(ctx, ValAddrs[0])
	require.Equal(t, amt.Sub(slashFraction.MulInt(amt).TruncateInt()), validator.GetBondedTokens())
	require.True(t, validator.IsJailed())

	// Case 3, slash unbonded validator
	validator, _ = keepers.StakingKeeper.GetValidator(ctx, ValAddrs[0])
	validator.Status = stakingtypes.Unbonded
	validator.Jailed = false
	validator.Tokens = amt
	keepers.StakingKeeper.SetValidator(ctx, validator)

	keepers.OracleKeeper.MissCounters.Insert(ctx, ValAddrs[0], uint64(votePeriodsPerWindow-minValidVotes+1))
	keepers.OracleKeeper.SlashAndResetMissCounters(ctx)
	validator, _ = keepers.StakingKeeper.GetValidator(ctx, ValAddrs[0])
	require.Equal(t, amt, validator.Tokens)
	require.False(t, validator.IsJailed())

	// Case 4, slash jailed validator
	validator, _ = keepers.StakingKeeper.GetValidator(ctx, ValAddrs[0])
	validator.Status = stakingtypes.Bonded
	validator.Jailed = true
	validator.Tokens = amt
	keepers.StakingKeeper.SetValidator(ctx, validator)

	keepers.OracleKeeper.MissCounters.Insert(ctx, ValAddrs[0], uint64(votePeriodsPerWindow-minValidVotes+1))
	keepers.OracleKeeper.SlashAndResetMissCounters(ctx)
	validator, _ = keepers.StakingKeeper.GetValidator(ctx, ValAddrs[0])
	require.Equal(t, amt, validator.Tokens)
}

func TestInvalidVotesSlashing(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1,
		e2eTesting.WithValidatorsNum(5),
		e2eTesting.WithBondAmount(testStakingAmt.String()),
	)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()
	msgServer := keeper.NewMsgServerImpl(keepers.OracleKeeper)

	vals := chain.GetCurrentValSet().Validators
	ValAddrs := make([]sdk.ValAddress, len(vals))
	for i := range vals {
		ValAddrs[i] = sdk.ValAddress(vals[i].Address)
	}

	params, err := keepers.OracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	params.VotePeriod = 1
	params.Whitelist = []asset.Pair{asset.Registry.Pair(denoms.ATOM, denoms.USD)}
	keepers.OracleKeeper.Params.Set(ctx, params)
	keepers.OracleKeeper.WhitelistedPairs.Insert(ctx, asset.Registry.Pair(denoms.ATOM, denoms.USD))

	votePeriodsPerWindow := math.LegacyNewDec(int64(keepers.OracleKeeper.SlashWindow(ctx))).QuoInt64(int64(keepers.OracleKeeper.VotePeriod(ctx))).TruncateInt64()
	slashFraction := keepers.OracleKeeper.SlashFraction(ctx)
	minValidPerWindow := keepers.OracleKeeper.MinValidPerWindow(ctx)

	for i := uint64(0); i < uint64(math.LegacyOneDec().Sub(minValidPerWindow).MulInt64(votePeriodsPerWindow).TruncateInt64()); i++ {
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

		// Account 1, govstable
		MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
			{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: testExchangeRate},
		}, vals[0])

		// Account 2, govstable, miss vote
		MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
			{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: testExchangeRate.Add(math.LegacyNewDec(100000000000000))},
		}, vals[1])

		// Account 3, govstable
		MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
			{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: testExchangeRate},
		}, vals[2])

		// Account 4, govstable
		MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
			{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: testExchangeRate},
		}, vals[3])

		keepers.OracleKeeper.UpdateExchangeRates(ctx)
		// keepers.OracleKeeper.SlashAndResetMissCounters(ctx)
		// keepers.OracleKeeper.UpdateExchangeRates(ctx)

		require.Equal(t, i+1, keepers.OracleKeeper.MissCounters.GetOr(ctx, ValAddrs[1], 0))
	}

	validator, err := keepers.StakingKeeper.GetValidator(ctx, ValAddrs[1])
	require.Equal(t, testStakingAmt, validator.GetBondedTokens())

	// one more miss vote will inccur ValAddrs[1] slashing
	// Account 1, govstable
	MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
		{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: testExchangeRate},
	}, vals[0])

	// Account 2, govstable, miss vote
	MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
		{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: testExchangeRate.Add(math.LegacyNewDec(100000000000000))},
	}, vals[1])

	// Account 3, govstable
	MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
		{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: testExchangeRate},
	}, vals[2])

	// Account 4, govstable
	MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
		{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: testExchangeRate},
	}, vals[3])

	ctx = ctx.WithBlockHeight(votePeriodsPerWindow - 1)
	keepers.OracleKeeper.UpdateExchangeRates(ctx)
	keepers.OracleKeeper.SlashAndResetMissCounters(ctx)
	// keepers.OracleKeeper.UpdateExchangeRates(ctx)

	validator, err = keepers.StakingKeeper.GetValidator(ctx, ValAddrs[1])
	require.NoError(t, err)
	require.Equal(t, math.LegacyOneDec().Sub(slashFraction).MulInt(testStakingAmt).TruncateInt(), validator.GetBondedTokens())
}

// TestWhitelistSlashing: Creates a scenario where one valoper (valIdx 0) does
// not vote throughout an entire vote window, while valopers 1 and 2 do.
func TestWhitelistSlashing(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1,
		e2eTesting.WithValidatorsNum(5),
		e2eTesting.WithBondAmount(testStakingAmt.String()),
	)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()
	msgServer := keeper.NewMsgServerImpl(keepers.OracleKeeper)

	vals := chain.GetCurrentValSet().Validators
	ValAddrs := make([]sdk.ValAddress, len(vals))
	for i := range vals {
		ValAddrs[i] = sdk.ValAddress(vals[i].Address)
	}

	params, err := keepers.OracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	params.VotePeriod = 1
	keepers.OracleKeeper.Params.Set(ctx, params)

	votePeriodsPerSlashWindow := math.LegacyNewDec(int64(keepers.OracleKeeper.SlashWindow(ctx))).QuoInt64(int64(keepers.OracleKeeper.VotePeriod(ctx))).TruncateInt64()
	minValidVotePeriodsPerWindow := keepers.OracleKeeper.MinValidPerWindow(ctx)

	pair := asset.Registry.Pair(denoms.ATOM, denoms.USD)
	priceVoteFromVal := func(valIdx int, block int64, erate math.LegacyDec) {
		MakeAggregatePrevoteAndVote(t, ctx, msgServer, block,
			types.ExchangeRateTuples{{Pair: pair, ExchangeRate: erate}},
			vals[valIdx])
	}
	keepers.OracleKeeper.WhitelistedPairs.Insert(ctx, pair)
	perfs := keepers.OracleKeeper.UpdateExchangeRates(ctx)
	require.EqualValues(t, 0, perfs.TotalRewardWeight())

	allowedMissPct := math.LegacyOneDec().Sub(minValidVotePeriodsPerWindow)
	allowedMissVotePeriods := allowedMissPct.MulInt64(votePeriodsPerSlashWindow).
		TruncateInt64()
	t.Logf("For %v blocks, valoper0 does not vote, while 1 and 2 do.", allowedMissVotePeriods)
	for idxMissPeriod := uint64(0); idxMissPeriod < uint64(allowedMissVotePeriods); idxMissPeriod++ {
		block := ctx.BlockHeight() + 1
		ctx = ctx.WithBlockHeight(block)

		valIdx := 0 // Valoper doesn't vote (abstain)
		priceVoteFromVal(valIdx+1, block, testExchangeRate)
		priceVoteFromVal(valIdx+2, block, testExchangeRate)

		perfs := keepers.OracleKeeper.UpdateExchangeRates(ctx)
		missCount := keepers.OracleKeeper.MissCounters.GetOr(ctx, ValAddrs[0], 0)
		require.EqualValues(t, 0, missCount, perfs.String())
	}

	t.Log("valoper0 should not be slashed")
	validator, err := keepers.StakingKeeper.Validator(ctx, ValAddrs[0])
	require.NoError(t, err)
	require.Equal(t, testStakingAmt, validator.GetBondedTokens())
}

func TestNotPassedBallotSlashing(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1, e2eTesting.WithValidatorsNum(5))
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()
	msgServer := keeper.NewMsgServerImpl(keepers.OracleKeeper)

	vals := chain.GetCurrentValSet().Validators
	ValAddrs := make([]sdk.ValAddress, len(vals))
	for i := range vals {
		ValAddrs[i] = sdk.ValAddress(vals[i].Address)
	}

	params, err := keepers.OracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	params.VotePeriod = 1
	params.Whitelist = []asset.Pair{asset.Registry.Pair(denoms.ATOM, denoms.USD)}
	keepers.OracleKeeper.Params.Set(ctx, params)

	// clear tobin tax to reset vote targets
	for _, p := range keepers.OracleKeeper.WhitelistedPairs.Iterate(ctx, collections.Range[asset.Pair]{}).Keys() {
		keepers.OracleKeeper.WhitelistedPairs.Delete(ctx, p)
	}
	keepers.OracleKeeper.WhitelistedPairs.Insert(ctx, asset.Registry.Pair(denoms.ATOM, denoms.USD))

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// Account 1, govstable
	MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: testExchangeRate}}, vals[0])

	keepers.OracleKeeper.UpdateExchangeRates(ctx)
	keepers.OracleKeeper.SlashAndResetMissCounters(ctx)
	// keepers.OracleKeeper.UpdateExchangeRates(ctx)
	require.Equal(t, uint64(0), keepers.OracleKeeper.MissCounters.GetOr(ctx, ValAddrs[0], 0))
	require.Equal(t, uint64(0), keepers.OracleKeeper.MissCounters.GetOr(ctx, ValAddrs[1], 0))
	require.Equal(t, uint64(0), keepers.OracleKeeper.MissCounters.GetOr(ctx, ValAddrs[2], 0))
}

func TestAbstainSlashing(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1,
		e2eTesting.WithValidatorsNum(5),
		e2eTesting.WithBondAmount(testStakingAmt.String()),
	)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()
	msgServer := keeper.NewMsgServerImpl(keepers.OracleKeeper)

	vals := chain.GetCurrentValSet().Validators
	ValAddrs := make([]sdk.ValAddress, len(vals))
	for i := range vals {
		ValAddrs[i] = sdk.ValAddress(vals[i].Address)
	}

	params, err := keepers.OracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	params.VotePeriod = 1
	// reset whitelisted pairs
	params.Whitelist = []asset.Pair{asset.Registry.Pair(denoms.ATOM, denoms.USD)}
	keepers.OracleKeeper.Params.Set(ctx, params)

	for _, p := range keepers.OracleKeeper.WhitelistedPairs.Iterate(ctx, collections.Range[asset.Pair]{}).Keys() {
		keepers.OracleKeeper.WhitelistedPairs.Delete(ctx, p)
	}
	keepers.OracleKeeper.WhitelistedPairs.Insert(ctx, asset.Registry.Pair(denoms.ATOM, denoms.USD))

	votePeriodsPerWindow := math.LegacyNewDec(int64(keepers.OracleKeeper.SlashWindow(ctx))).QuoInt64(int64(keepers.OracleKeeper.VotePeriod(ctx))).TruncateInt64()
	minValidPerWindow := keepers.OracleKeeper.MinValidPerWindow(ctx)

	for i := uint64(0); i <= uint64(math.LegacyOneDec().Sub(minValidPerWindow).MulInt64(votePeriodsPerWindow).TruncateInt64()); i++ {
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

		// Account 1, ATOM/USD
		MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: testExchangeRate}}, vals[0])

		// Account 2, ATOM/USD, abstain vote
		MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: math.LegacyOneDec().Neg()}}, vals[1])

		// Account 3, ATOM/USD
		MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{{Pair: asset.Registry.Pair(denoms.ATOM, denoms.USD), ExchangeRate: testExchangeRate}}, vals[2])

		keepers.OracleKeeper.UpdateExchangeRates(ctx)
		keepers.OracleKeeper.SlashAndResetMissCounters(ctx)
		// keepers.OracleKeeper.UpdateExchangeRates(ctx)
		require.Equal(t, uint64(0), keepers.OracleKeeper.MissCounters.GetOr(ctx, ValAddrs[1], 0))
	}

	validator, err := keepers.StakingKeeper.Validator(ctx, ValAddrs[1])
	require.NoError(t, err)
	require.Equal(t, testStakingAmt, validator.GetBondedTokens())
}
