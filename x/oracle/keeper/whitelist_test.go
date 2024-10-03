package keeper_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/NibiruChain/collections"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/common/asset"
	"github.com/archway-network/archway/x/common/denoms"
	"github.com/archway-network/archway/x/common/set"
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

			for _, p := range keepers.OracleKeeper.WhitelistedPairs.Iterate(ctx, collections.Range[asset.Pair]{}).Keys() {
				keepers.OracleKeeper.WhitelistedPairs.Delete(ctx, p)
			}

			expectedTargets := tc.in
			for _, target := range expectedTargets {
				keepers.OracleKeeper.WhitelistedPairs.Insert(ctx, target)
			}

			var panicAssertFn func(t assert.TestingT, f assert.PanicTestFunc, msgAndArgs ...interface{}) bool
			switch tc.panic {
			case true:
				panicAssertFn = assert.Panics
			default:
				panicAssertFn = assert.NotPanics
			}
			panicAssertFn(t, func() {
				targets := keepers.OracleKeeper.GetWhitelistedPairs(ctx)
				assert.Equal(t, expectedTargets, targets)
			})
		})
	}

	chain := e2eTesting.NewTestChain(t, 1)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()

	for _, p := range keepers.OracleKeeper.WhitelistedPairs.Iterate(ctx, collections.Range[asset.Pair]{}).Keys() {
		keepers.OracleKeeper.WhitelistedPairs.Delete(ctx, p)
	}

	expectedTargets := []asset.Pair{"foo:bar", "whoo:whoo"}
	for _, target := range expectedTargets {
		keepers.OracleKeeper.WhitelistedPairs.Insert(ctx, target)
	}

	targets := keepers.OracleKeeper.GetWhitelistedPairs(ctx)
	require.Equal(t, expectedTargets, targets)
}

func TestIsWhitelistedPair(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()

	for _, p := range keepers.OracleKeeper.WhitelistedPairs.Iterate(ctx, collections.Range[asset.Pair]{}).Keys() {
		keepers.OracleKeeper.WhitelistedPairs.Delete(ctx, p)
	}

	validPairs := []asset.Pair{"foo:bar", "xxx:yyy", "whoo:whoo"}
	for _, target := range validPairs {
		keepers.OracleKeeper.WhitelistedPairs.Insert(ctx, target)
		require.True(t, keepers.OracleKeeper.IsWhitelistedPair(ctx, target))
	}
}

func TestUpdateWhitelist(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	keepers := chain.GetApp().Keepers
	ctx := chain.GetContext()

	// prepare test by resetting the genesis pairs
	for _, p := range keepers.OracleKeeper.WhitelistedPairs.Iterate(ctx, collections.Range[asset.Pair]{}).Keys() {
		keepers.OracleKeeper.WhitelistedPairs.Delete(ctx, p)
	}

	currentWhitelist := set.New(asset.NewPair(denoms.NIBI, denoms.USD), asset.NewPair(denoms.BTC, denoms.USD))
	for p := range currentWhitelist {
		keepers.OracleKeeper.WhitelistedPairs.Insert(ctx, p)
	}

	nextWhitelist := set.New(asset.NewPair(denoms.NIBI, denoms.USD), asset.NewPair(denoms.BTC, denoms.USD))

	// no updates case
	whitelistSlice := nextWhitelist.ToSlice()
	sort.Slice(whitelistSlice, func(i, j int) bool {
		return whitelistSlice[i].String() < whitelistSlice[j].String()
	})
	keepers.OracleKeeper.RefreshWhitelist(ctx, whitelistSlice, currentWhitelist)
	assert.Equal(t, whitelistSlice, keepers.OracleKeeper.GetWhitelistedPairs(ctx))

	// len update (fast path)
	nextWhitelist.Add(asset.NewPair(denoms.NIBI, denoms.ETH))
	whitelistSlice = nextWhitelist.ToSlice()
	sort.Slice(whitelistSlice, func(i, j int) bool {
		return whitelistSlice[i].String() < whitelistSlice[j].String()
	})
	keepers.OracleKeeper.RefreshWhitelist(ctx, whitelistSlice, currentWhitelist)
	assert.Equal(t, whitelistSlice, keepers.OracleKeeper.GetWhitelistedPairs(ctx))

	// diff update (slow path)
	currentWhitelist.Add(asset.NewPair(denoms.NIBI, denoms.ATOM))
	whitelistSlice = nextWhitelist.ToSlice()
	sort.Slice(whitelistSlice, func(i, j int) bool {
		return whitelistSlice[i].String() < whitelistSlice[j].String()
	})
	keepers.OracleKeeper.RefreshWhitelist(ctx, whitelistSlice, currentWhitelist)
	assert.Equal(t, whitelistSlice, keepers.OracleKeeper.GetWhitelistedPairs(ctx))
}
