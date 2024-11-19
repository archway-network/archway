package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/oracle/asset"
	"github.com/archway-network/archway/x/oracle/denoms"
	"github.com/archway-network/archway/x/oracle/types"
)

func TestParams(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()

	// Test default params setting
	require.NoError(t, keepers.OracleKeeper.Params.Set(ctx, types.DefaultParams()))
	params, err := keepers.OracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	require.NotNil(t, params)

	// Test custom params setting
	votePeriod := uint64(10)
	voteThreshold := math.LegacyNewDecWithPrec(33, 2)
	minVoters := uint64(4)
	oracleRewardBand := math.LegacyNewDecWithPrec(1, 2)
	slashFraction := math.LegacyNewDecWithPrec(1, 2)
	slashWindow := uint64(1000)
	minValidPerWindow := math.LegacyNewDecWithPrec(1, 4)
	minFeeRatio := math.LegacyNewDecWithPrec(1, 2)
	whitelist := []asset.Pair{
		asset.Registry.Pair(denoms.BTC, denoms.NUSD),
		asset.Registry.Pair(denoms.ETH, denoms.NUSD),
	}

	// Should really test validateParams, but skipping because obvious
	newParams := types.Params{
		VotePeriod:        votePeriod,
		VoteThreshold:     voteThreshold,
		MinVoters:         minVoters,
		RewardBand:        oracleRewardBand,
		Whitelist:         whitelist,
		SlashFraction:     slashFraction,
		SlashWindow:       slashWindow,
		MinValidPerWindow: minValidPerWindow,
		ValidatorFeeRatio: minFeeRatio,
	}
	require.NoError(t, keepers.OracleKeeper.Params.Set(ctx, newParams))

	storedParams, err := keepers.OracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	require.NotNil(t, storedParams)
	require.Equal(t, storedParams, newParams)
}
