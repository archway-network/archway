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
	accNum := 3

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

	// test self delegation
	require.NoError(t, keepers.OracleKeeper.ValidateFeeder(ctx, sdk.AccAddress(ValAddrs[0]), ValAddrs[0]))
	require.NoError(t, keepers.OracleKeeper.ValidateFeeder(ctx, sdk.AccAddress(ValAddrs[1]), ValAddrs[1]))

	// delegate works
	keepers.OracleKeeper.FeederDelegations.Set(ctx, ValAddrs[0], AccAddrs[1])
	require.NoError(t, keepers.OracleKeeper.ValidateFeeder(ctx, AccAddrs[1], ValAddrs[0]))
	require.Error(t, keepers.OracleKeeper.ValidateFeeder(ctx, AccAddrs[2], ValAddrs[0]))

	// only active validators can do oracle votes
	validator.Status = stakingtypes.Unbonded
	keepers.StakingKeeper.SetValidator(ctx, validator)
	require.Error(t, keepers.OracleKeeper.ValidateFeeder(ctx, AccAddrs[1], ValAddrs[0]))
}
