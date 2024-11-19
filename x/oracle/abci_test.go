package oracle_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	cmTypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/oracle"
	"github.com/archway-network/archway/x/oracle/asset"
	"github.com/archway-network/archway/x/oracle/denoms"
	"github.com/archway-network/archway/x/oracle/keeper"
	"github.com/archway-network/archway/x/oracle/types"
	oracletypes "github.com/archway-network/archway/x/oracle/types"
)

// TODO (spekalsg3): duplicated from `package keeper_test`
func MakeAggregatePrevoteAndVote(
	t *testing.T,
	ctx sdk.Context,
	msgServer oracletypes.MsgServer,
	height int64,
	rates oracletypes.ExchangeRateTuples,
	val *cmTypes.Validator,
) {
	accAddr := sdk.AccAddress(val.Address)
	valAddr := sdk.ValAddress(val.Address)

	salt := "1"
	ratesStr, err := rates.ToString()
	require.NoError(t, err)
	hash := oracletypes.GetAggregateVoteHash(salt, ratesStr, valAddr)

	prevoteMsg := oracletypes.NewMsgAggregateExchangeRatePrevote(hash, accAddr, valAddr)
	_, err = msgServer.AggregateExchangeRatePrevote(ctx.WithBlockHeight(height), prevoteMsg)
	require.NoError(t, err)

	// chain.GetApp().Keepers.OracleKeeper.VotePeriod(ctx)
	voteMsg := oracletypes.NewMsgAggregateExchangeRateVote(salt, ratesStr, accAddr, valAddr)
	_, err = msgServer.AggregateExchangeRateVote(ctx.WithBlockHeight(height+1), voteMsg)
	require.NoError(t, err)
}

func TestOracleTallyTiming(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1,
		e2eTesting.WithValidatorsNum(4),
	)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()
	msgServer := keeper.NewMsgServerImpl(keepers.OracleKeeper)

	params, err := keepers.OracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	params.VotePeriod = 1
	require.NoError(t, keepers.OracleKeeper.Params.Set(ctx, params))

	// all the Addrs vote for the block ... not last period block yet, so tally fails
	for _, val := range chain.GetCurrentValSet().Validators {
		MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
			{Pair: asset.Registry.Pair(denoms.BTC, denoms.USD), ExchangeRate: math.LegacyOneDec()},
		}, val)
	}

	params.VotePeriod = 10 // set vote period to 10 for now, for convenience
	params.ExpirationBlocks = 100
	require.NoError(t, keepers.OracleKeeper.Params.Set(ctx, params))
	require.Equal(t, 1, int(ctx.BlockHeight()))

	require.NoError(t, oracle.EndBlocker(ctx, keepers.OracleKeeper))
	_, err = keepers.OracleKeeper.ExchangeRates.Get(ctx, asset.Registry.Pair(denoms.BTC, denoms.USD))
	require.Error(t, err)

	ctx = ctx.WithBlockHeight(int64(params.VotePeriod))
	require.NoError(t, oracle.EndBlocker(ctx, keepers.OracleKeeper))

	_, err = keepers.OracleKeeper.ExchangeRates.Get(ctx, asset.Registry.Pair(denoms.BTC, denoms.USD))
	require.NoError(t, err)
}

// Set prices for 2 pairs, one that is updated and the other which is updated only once.
// Ensure that the updated pair is not deleted and the other pair is deleted after a certain time.
func TestOraclePriceExpiration(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1,
		e2eTesting.WithValidatorsNum(4),
	)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()
	msgServer := keeper.NewMsgServerImpl(keepers.OracleKeeper)
	pair1 := asset.Registry.Pair(denoms.BTC, denoms.USD)
	pair2 := asset.Registry.Pair(denoms.ETH, denoms.USD)

	params, err := keepers.OracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	params.VotePeriod = 1
	require.NoError(t, keepers.OracleKeeper.Params.Set(ctx, params))

	// Set prices for both pairs
	for _, val := range chain.GetCurrentValSet().Validators {
		MakeAggregatePrevoteAndVote(t, ctx, msgServer, 0, types.ExchangeRateTuples{
			{Pair: pair1, ExchangeRate: math.LegacyOneDec()},
			{Pair: pair2, ExchangeRate: math.LegacyOneDec()},
		}, val)
	}

	params.VotePeriod = 10
	params.ExpirationBlocks = 10
	require.NoError(t, keepers.OracleKeeper.Params.Set(ctx, params))

	// Wait for prices to set
	ctx = ctx.WithBlockHeight(int64(params.VotePeriod))
	require.NoError(t, oracle.EndBlocker(ctx, keepers.OracleKeeper))

	// Check if both prices are set
	_, err = keepers.OracleKeeper.ExchangeRates.Get(ctx, pair1)
	require.NoError(t, err)
	_, err = keepers.OracleKeeper.ExchangeRates.Get(ctx, pair2)
	require.NoError(t, err)

	// Set prices for pair 1
	voteHeight := int64(params.VotePeriod+params.ExpirationBlocks) - 1
	for _, val := range chain.GetCurrentValSet().Validators {
		MakeAggregatePrevoteAndVote(t, ctx, msgServer, voteHeight, types.ExchangeRateTuples{
			{Pair: pair1, ExchangeRate: math.LegacyNewDec(2)},
		}, val)
	}

	// Set price
	ctx = ctx.WithBlockHeight(voteHeight)
	require.NoError(t, oracle.EndBlocker(ctx, keepers.OracleKeeper))

	// Set the block height to the expiration height
	// End blocker should delete the price of pair2
	ctx = ctx.WithBlockHeight(int64(params.ExpirationBlocks + params.VotePeriod))
	require.NoError(t, oracle.EndBlocker(ctx, keepers.OracleKeeper))

	_, err = keepers.OracleKeeper.ExchangeRates.Get(ctx, pair1)
	require.NoError(t, err)
	_, err = keepers.OracleKeeper.ExchangeRates.Get(ctx, pair2)
	require.Error(t, err)
}
