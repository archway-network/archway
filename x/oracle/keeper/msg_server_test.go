package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/oracle/asset"
	"github.com/archway-network/archway/x/oracle/denoms"
	"github.com/archway-network/archway/x/oracle/keeper"
	"github.com/archway-network/archway/x/oracle/types"
)

func TestFeederDelegation(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1,
		e2eTesting.WithValidatorsNum(3),
	)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()
	msgServer := keeper.NewMsgServerImpl(keepers.OracleKeeper)

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
	keepers.OracleKeeper.Params.Set(ctx, params)

	exchangeRates := types.ExchangeRateTuples{
		{
			Pair:         asset.Registry.Pair(denoms.BTC, denoms.USD),
			ExchangeRate: testExchangeRate,
		},
	}

	exchangeRateStr, err := exchangeRates.ToString()
	require.NoError(t, err)
	salt := "1"
	hash := types.GetAggregateVoteHash(salt, exchangeRateStr, ValAddrs[0])

	// Case 1: empty message
	delegateFeedConsentMsg := types.MsgDelegateFeedConsent{}
	_, err = msgServer.DelegateFeedConsent(ctx, &delegateFeedConsentMsg)
	require.Error(t, err)

	// Case 2: Normal Prevote - without delegation
	prevoteMsg := types.NewMsgAggregateExchangeRatePrevote(hash, AccAddrs[0], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRatePrevote(ctx, prevoteMsg)
	require.NoError(t, err)

	// Case 2.1: Normal Prevote - with delegation fails
	prevoteMsg = types.NewMsgAggregateExchangeRatePrevote(hash, AccAddrs[1], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRatePrevote(ctx, prevoteMsg)
	require.Error(t, err)

	// Case 2.2: Normal Vote - without delegation
	voteMsg := types.NewMsgAggregateExchangeRateVote(salt, exchangeRateStr, AccAddrs[0], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(ctx.WithBlockHeight(2), voteMsg)
	require.NoError(t, err)

	// Case 2.3: Normal Vote - with delegation fails
	voteMsg = types.NewMsgAggregateExchangeRateVote(salt, exchangeRateStr, AccAddrs[1], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(ctx.WithBlockHeight(2), voteMsg)
	require.Error(t, err)

	// Case 3: Normal MsgDelegateFeedConsent succeeds
	msg := types.NewMsgDelegateFeedConsent(ValAddrs[0], AccAddrs[1])
	_, err = msgServer.DelegateFeedConsent(ctx, msg)
	require.NoError(t, err)

	// Case 4.1: Normal Prevote - without delegation fails
	prevoteMsg = types.NewMsgAggregateExchangeRatePrevote(hash, AccAddrs[2], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRatePrevote(ctx, prevoteMsg)
	require.Error(t, err)

	// Case 4.2: Normal Prevote - with delegation succeeds
	prevoteMsg = types.NewMsgAggregateExchangeRatePrevote(hash, AccAddrs[1], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRatePrevote(ctx, prevoteMsg)
	require.NoError(t, err)

	// Case 4.3: Normal Vote - without delegation fails
	voteMsg = types.NewMsgAggregateExchangeRateVote(salt, exchangeRateStr, AccAddrs[2], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(ctx.WithBlockHeight(2), voteMsg)
	require.Error(t, err)

	// Case 4.4: Normal Vote - with delegation succeeds
	voteMsg = types.NewMsgAggregateExchangeRateVote(salt, exchangeRateStr, AccAddrs[1], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(ctx.WithBlockHeight(2), voteMsg)
	require.NoError(t, err)
}

func TestAggregatePrevoteVote(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1, e2eTesting.WithValidatorsNum(2))
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()
	msgServer := keeper.NewMsgServerImpl(keepers.OracleKeeper)

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
	keepers.OracleKeeper.Params.Set(ctx, params)

	salt := "1"
	exchangeRates := types.ExchangeRateTuples{
		{
			Pair:         asset.Registry.Pair(denoms.ATOM, denoms.USD),
			ExchangeRate: math.LegacyMustNewDecFromStr("1000.23"),
		},
		{
			Pair:         asset.Registry.Pair(denoms.ETH, denoms.USD),
			ExchangeRate: math.LegacyMustNewDecFromStr("0.29"),
		},

		{
			Pair:         asset.Registry.Pair(denoms.BTC, denoms.USD),
			ExchangeRate: math.LegacyMustNewDecFromStr("0.27"),
		},
	}

	otherExchangeRate := types.ExchangeRateTuples{
		{
			Pair:         asset.Registry.Pair(denoms.ATOM, denoms.USD),
			ExchangeRate: math.LegacyMustNewDecFromStr("1000.23"),
		},
		{
			Pair:         asset.Registry.Pair(denoms.ETH, denoms.USD),
			ExchangeRate: math.LegacyMustNewDecFromStr("0.29"),
		},

		{
			Pair:         asset.Registry.Pair(denoms.ETH, denoms.USD),
			ExchangeRate: math.LegacyMustNewDecFromStr("0.27"),
		},
	}

	unintendedExchangeRateStr := types.ExchangeRateTuples{
		{
			Pair:         asset.Registry.Pair(denoms.ATOM, denoms.USD),
			ExchangeRate: math.LegacyMustNewDecFromStr("1000.23"),
		},
		{
			Pair:         asset.Registry.Pair(denoms.ETH, denoms.USD),
			ExchangeRate: math.LegacyMustNewDecFromStr("0.29"),
		},
		{
			Pair:         "BTC:CNY",
			ExchangeRate: math.LegacyMustNewDecFromStr("0.27"),
		},
	}
	exchangeRatesStr, err := exchangeRates.ToString()
	require.NoError(t, err)

	otherExchangeRateStr, err := otherExchangeRate.ToString()
	require.NoError(t, err)

	unintendedExchageRateStr, err := unintendedExchangeRateStr.ToString()
	require.NoError(t, err)

	hash := types.GetAggregateVoteHash(salt, exchangeRatesStr, ValAddrs[0])

	aggregateExchangeRatePrevoteMsg := types.NewMsgAggregateExchangeRatePrevote(hash, AccAddrs[0], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRatePrevote(ctx, aggregateExchangeRatePrevoteMsg)
	require.NoError(t, err)

	// Unauthorized feeder
	aggregateExchangeRatePrevoteMsg = types.NewMsgAggregateExchangeRatePrevote(hash, AccAddrs[1], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRatePrevote(ctx, aggregateExchangeRatePrevoteMsg)
	require.Error(t, err)

	// Invalid addr
	aggregateExchangeRatePrevoteMsg = types.NewMsgAggregateExchangeRatePrevote(hash, sdk.AccAddress{}, ValAddrs[0])
	_, err = msgServer.AggregateExchangeRatePrevote(ctx, aggregateExchangeRatePrevoteMsg)
	require.Error(t, err)

	// Invalid validator addr
	aggregateExchangeRatePrevoteMsg = types.NewMsgAggregateExchangeRatePrevote(hash, AccAddrs[0], sdk.ValAddress{})
	_, err = msgServer.AggregateExchangeRatePrevote(ctx, aggregateExchangeRatePrevoteMsg)
	require.Error(t, err)

	// Invalid reveal period
	aggregateExchangeRateVoteMsg := types.NewMsgAggregateExchangeRateVote(salt, exchangeRatesStr, AccAddrs[0], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(ctx, aggregateExchangeRateVoteMsg)
	require.Error(t, err)

	// Invalid reveal period
	ctx = ctx.WithBlockHeight(3)
	aggregateExchangeRateVoteMsg = types.NewMsgAggregateExchangeRateVote(salt, exchangeRatesStr, AccAddrs[0], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(ctx, aggregateExchangeRateVoteMsg)
	require.Error(t, err)

	// Other exchange rate with valid real period
	ctx = ctx.WithBlockHeight(2)
	aggregateExchangeRateVoteMsg = types.NewMsgAggregateExchangeRateVote(salt, otherExchangeRateStr, AccAddrs[0], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(ctx, aggregateExchangeRateVoteMsg)
	require.Error(t, err)

	// Unauthorized feeder
	aggregateExchangeRateVoteMsg = types.NewMsgAggregateExchangeRateVote(salt, exchangeRatesStr, AccAddrs[1], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(ctx, aggregateExchangeRateVoteMsg)
	require.Error(t, err)

	// Unintended denom vote
	aggregateExchangeRateVoteMsg = types.NewMsgAggregateExchangeRateVote(salt, unintendedExchageRateStr, AccAddrs[0], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(ctx, aggregateExchangeRateVoteMsg)
	require.Error(t, err)

	// Valid exchange rate reveal submission
	ctx = ctx.WithBlockHeight(2)
	aggregateExchangeRateVoteMsg = types.NewMsgAggregateExchangeRateVote(salt, exchangeRatesStr, AccAddrs[0], ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(ctx, aggregateExchangeRateVoteMsg)
	require.NoError(t, err)
}
