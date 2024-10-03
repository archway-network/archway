package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/NibiruChain/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/common"
	"github.com/archway-network/archway/x/common/asset"
	"github.com/archway-network/archway/x/common/denoms"
	"github.com/archway-network/archway/x/oracle/keeper"
	"github.com/archway-network/archway/x/oracle/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

func TestKeeperRewardsDistributionMultiVotePeriods(t *testing.T) {
	// this simulates allocating rewards for the pair atom:usd
	// over 5 voting periods. It simulates rewards are correctly
	// distributed over 5 voting periods to 5 validators.
	// then we simulate that after the 5 voting periods are
	// finished no more rewards distribution happen.
	const periods uint64 = 5
	const validators = 5

	chain := e2eTesting.NewTestChain(t, 1, e2eTesting.WithValidatorsNum(validators))
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()

	params, err := keepers.OracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	params.VotePeriod = 1
	keepers.OracleKeeper.Params.Set(ctx, params)

	msgServer := keeper.NewMsgServerImpl(keepers.OracleKeeper)

	vals := chain.GetCurrentValSet().Validators

	rewards := sdk.NewInt64Coin("reward", 1*common.TO_MICRO)
	valPeriodicRewards := sdk.NewDecCoinsFromCoins(rewards).
		QuoDec(math.LegacyNewDec(int64(periods))).
		QuoDec(math.LegacyNewDec(int64(validators)))
	require.NoError(t, keepers.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(rewards)))
	require.NoError(t, keepers.OracleKeeper.AllocateRewards(ctx, minttypes.ModuleName, sdk.NewCoins(rewards), periods))

	for i := uint64(1); i <= periods; i++ {
		for _, val := range vals {
			// for doc's sake, this function is capable of making prevotes and votes because it
			// passes the current context block height for pre vote
			// then changes the height to current height + vote period for the vote
			MakeAggregatePrevoteAndVote(t, ctx, msgServer, ctx.BlockHeight(), types.ExchangeRateTuples{
				{
					Pair:         asset.Registry.Pair(denoms.ATOM, denoms.USD),
					ExchangeRate: testExchangeRate,
				},
			}, val)
		}

		keepers.OracleKeeper.UpdateExchangeRates(ctx)

		for valIndex := 0; valIndex < validators; valIndex++ {
			distributionRewards, err := keepers.DistrKeeper.GetValidatorOutstandingRewards(ctx, sdk.ValAddress(vals[0].Address))
			require.NoError(t, err)
			truncatedGot, _ := distributionRewards.Rewards.
				QuoDec(math.LegacyNewDec(int64(i))). // outstanding rewards will count for the previous vote period too, so we divide it by current period
				TruncateDecimal()                    // NOTE: not applying this on truncatedExpected because of rounding the test fails
			truncatedExpected, _ := valPeriodicRewards.TruncateDecimal()

			require.Equalf(t, truncatedExpected, truncatedGot, "period: %d, %s <-> %s", i, truncatedExpected.String(), truncatedGot.String())
		}
		// assert rewards

		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + int64(params.VotePeriod))
	}

	// assert there are no rewards
	require.True(t, keepers.OracleKeeper.GatherRewardsForVotePeriod(ctx).IsZero())

	// assert that there are no rewards instances
	require.Empty(t, keepers.OracleKeeper.Rewards.Iterate(ctx, collections.Range[uint64]{}).Keys())
}
