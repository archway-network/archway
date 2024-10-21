package keeper_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/types/set"
	"github.com/archway-network/archway/x/oracle/asset"
	"github.com/archway-network/archway/x/oracle/denoms"
)

func TestKeeper_GetVoteTargets(t *testing.T) {
	type TestCase struct {
		name  string
		in    []asset.Pair
		panic bool
	}

	panicCases := []TestCase{
		{name: "blank pair", in: []asset.Pair{""}, panic: true},
		{name: "blank pair and others", in: []asset.Pair{"", "x", "abc", "defafask"}, panic: true},
		{name: "denom len too short", in: []asset.Pair{"x:y", "xx:yy"}, panic: true},
	}
	happyCases := []TestCase{
		{name: "happy", in: []asset.Pair{"foo:bar", "whoo:whoo"}},
	}

	for _, testCase := range append(panicCases, happyCases...) {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			chain := e2eTesting.NewTestChain(t, 1)
			keepers := chain.GetApp().Keepers
			ctx := chain.GetContext()

			err := keepers.OracleKeeper.WhitelistedPairs.Clear(ctx, nil)
			require.NoError(t, err)

			expectedTargets := tc.in
			for _, target := range expectedTargets {
				keepers.OracleKeeper.WhitelistedPairs.Set(ctx, target)
			}

			var panicAssertFn func(t assert.TestingT, f assert.PanicTestFunc, msgAndArgs ...interface{}) bool
			switch tc.panic {
			case true:
				panicAssertFn = assert.Panics
			default:
				panicAssertFn = assert.NotPanics
			}
			panicAssertFn(t, func() {
				targets, err := keepers.OracleKeeper.GetWhitelistedPairs(ctx)
				require.NoError(t, err)
				assert.Equal(t, expectedTargets, targets)
			})
		})
	}

	chain := e2eTesting.NewTestChain(t, 1)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()

	err := keepers.OracleKeeper.WhitelistedPairs.Clear(ctx, nil)
	require.NoError(t, err)

	expectedTargets := []asset.Pair{"foo:bar", "whoo:whoo"}
	for _, target := range expectedTargets {
		keepers.OracleKeeper.WhitelistedPairs.Set(ctx, target)
	}

	targets, err := keepers.OracleKeeper.GetWhitelistedPairs(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedTargets, targets)
}

func TestIsWhitelistedPair(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()

	err := keepers.OracleKeeper.WhitelistedPairs.Clear(ctx, nil)
	require.NoError(t, err)

	validPairs := []asset.Pair{"foo:bar", "xxx:yyy", "whoo:whoo"}
	for _, target := range validPairs {
		keepers.OracleKeeper.WhitelistedPairs.Set(ctx, target)
		flag, err := keepers.OracleKeeper.IsWhitelistedPair(ctx, target)
		require.NoError(t, err)
		require.True(t, flag)
	}
}

func TestUpdateWhitelist(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()

	// prepare test by resetting the genesis pairs
	err := keepers.OracleKeeper.WhitelistedPairs.Clear(ctx, nil)
	require.NoError(t, err)

	currentWhitelist := set.New(asset.NewPair(denoms.NIBI, denoms.USD), asset.NewPair(denoms.BTC, denoms.USD))
	for p := range currentWhitelist {
		keepers.OracleKeeper.WhitelistedPairs.Set(ctx, p)
	}

	nextWhitelist := set.New(asset.NewPair(denoms.NIBI, denoms.USD), asset.NewPair(denoms.BTC, denoms.USD))

	// no updates case
	whitelistSlice := nextWhitelist.ToSlice()
	sort.Slice(whitelistSlice, func(i, j int) bool {
		return whitelistSlice[i].String() < whitelistSlice[j].String()
	})
	keepers.OracleKeeper.RefreshWhitelist(ctx, whitelistSlice, currentWhitelist)
	pairs, err := keepers.OracleKeeper.GetWhitelistedPairs(ctx)
	require.NoError(t, err)
	assert.Equal(t, whitelistSlice, pairs)

	// len update (fast path)
	nextWhitelist.Add(asset.NewPair(denoms.NIBI, denoms.ETH))
	whitelistSlice = nextWhitelist.ToSlice()
	sort.Slice(whitelistSlice, func(i, j int) bool {
		return whitelistSlice[i].String() < whitelistSlice[j].String()
	})
	keepers.OracleKeeper.RefreshWhitelist(ctx, whitelistSlice, currentWhitelist)
	pairs, err = keepers.OracleKeeper.GetWhitelistedPairs(ctx)
	require.NoError(t, err)
	assert.Equal(t, whitelistSlice, pairs)

	// diff update (slow path)
	currentWhitelist.Add(asset.NewPair(denoms.NIBI, denoms.ATOM))
	whitelistSlice = nextWhitelist.ToSlice()
	sort.Slice(whitelistSlice, func(i, j int) bool {
		return whitelistSlice[i].String() < whitelistSlice[j].String()
	})
	keepers.OracleKeeper.RefreshWhitelist(ctx, whitelistSlice, currentWhitelist)
	pairs, err = keepers.OracleKeeper.GetWhitelistedPairs(ctx)
	require.NoError(t, err)
	assert.Equal(t, whitelistSlice, pairs)
}
