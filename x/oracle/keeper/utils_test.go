// nolint
package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	e2eTesting "github.com/archway-network/archway/e2e/testing"
	oracletypes "github.com/archway-network/archway/x/oracle/types"

	cmTypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"
)

// Test addresses
var (
	InitTokens = sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	InitCoins  = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, InitTokens))

	testStakingAmt   = sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	testExchangeRate = math.LegacyNewDec(1700)

	OracleDecPrecision = 8
)

func AllocateRewards(t *testing.T, chain e2eTesting.TestChain, rewards sdk.Coins, votePeriods uint64) {
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()
	require.NoError(t, keepers.BankKeeper.MintCoins(ctx, minttypes.ModuleName, rewards))
	require.NoError(t, keepers.OracleKeeper.AllocateRewards(ctx, minttypes.ModuleName, rewards, votePeriods))
}

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
