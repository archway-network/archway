package keeper_test

import (
	"testing"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func TestValidateFeeder(t *testing.T) {
	// initial setup
	chain := e2eTesting.NewTestChain(t, 1,
		// e2eTesting.WithCallbackParams(123),
		e2eTesting.WithValidatorsNum(2),
	)
	keepers := chain.GetApp().Keepers
	amt := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
	ctx := chain.GetContext()

	vals := chain.GetCurrentValSet().Validators
	AccAddrs := make([]sdk.AccAddress, len(vals))
	ValAddrs := make([]sdk.ValAddress, len(vals))
	for i := range vals {
		AccAddrs[i] = sdk.AccAddress(vals[i].Address)
		ValAddrs[i] = sdk.ValAddress(vals[i].Address)
	}

	InitTokens := sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)

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
	require.Equal(t, amt, validator1.GetBondedTokens())

	require.NoError(t, keepers.OracleKeeper.ValidateFeeder(ctx, AccAddrs[0], ValAddrs[0]))
	require.NoError(t, keepers.OracleKeeper.ValidateFeeder(ctx, AccAddrs[1], ValAddrs[1]))

	// delegate works
	keepers.OracleKeeper.FeederDelegations.Insert(ctx, ValAddrs[0], AccAddrs[1])
	require.NoError(t, keepers.OracleKeeper.ValidateFeeder(ctx, AccAddrs[1], ValAddrs[0]))
	require.Error(t, keepers.OracleKeeper.ValidateFeeder(ctx, AccAddrs[2], ValAddrs[0]))

	// only active validators can do oracle votes
	validator.Status = stakingtypes.Unbonded
	keepers.StakingKeeper.SetValidator(ctx, validator)
	require.Error(t, keepers.OracleKeeper.ValidateFeeder(ctx, AccAddrs[1], ValAddrs[0]))
}
