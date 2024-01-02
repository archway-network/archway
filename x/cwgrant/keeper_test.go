package cwgrant_test

import (
	"encoding/json"
	"fmt"
	"testing"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/cwgrant/types"
)

func TestKeeper(t *testing.T) {
	app := e2eTesting.NewTestChain(t, 0)
	k := app.GetApp().Keepers.CWGrantKeeper
	ctx := app.GetContext()

	t.Run("register as granter – not a cw contract", func(t *testing.T) {
		acc := app.GetAccount(0)
		err := k.RegisterAsGranter(ctx, acc.Address)
		require.ErrorIs(t, err, types.ErrNotAContract)
	})

	t.Run("state import and export", func(t *testing.T) {
		wantState := &types.GenesisState{GrantingContracts: []string{sdk.AccAddress("alice").String(), sdk.AccAddress("bob").String()}}
		err := k.ImportState(ctx, wantState)
		require.NoError(t, err)

		gotState, err := k.ExportState(ctx)
		require.NoError(t, err)

		require.Equal(t, wantState, gotState)
	})
}

func TestFullIntegration(t *testing.T) {
	app := e2eTesting.NewTestChain(t, 0, e2eTesting.WithGenAccounts(10))
	deployer := app.GetAccount(0)

	codeID := app.UploadContract(deployer, "../../contracts/cwgrant/artifacts/cwgrant.wasm", wasmdTypes.DefaultUploadAccess)

	grantedAcc := app.GetAccount(1) // account who receives grants.
	initMsg := fmt.Sprintf(`{"grants": ["%s"]}`, grantedAcc.Address)
	cwGranter, _ := app.InstantiateContract(deployer, codeID, deployer.Address.String(), "cwgrant", sdk.NewCoins(sdk.NewInt64Coin("stake", 1_000_000_000_000)), json.RawMessage(initMsg))

	// register as cwgrant contract.
	err := app.GetApp().Keepers.CWGrantKeeper.RegisterAsGranter(app.GetContext(), cwGranter)
	require.NoError(t, err)

	// now try to send a tx with a cw granter
	msg := &banktypes.MsgSend{
		FromAddress: grantedAcc.Address.String(),
		ToAddress:   deployer.Address.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("stake", 1)),
	}

	grantedBalanceBefore := app.GetBalance(grantedAcc.Address)
	cwGranterBalanceBefore := app.GetBalance(cwGranter)
	fees := sdk.NewInt64Coin("stake", 100_000_000_000)
	_, _, _, err = app.SendMsgs(grantedAcc, true, []sdk.Msg{msg}, e2eTesting.WithGranter(cwGranter), e2eTesting.WithMsgFees(fees))
	require.NoError(t, err)

	grantedBalanceAfter := app.GetBalance(grantedAcc.Address)
	cwGranterBalanceAfter := app.GetBalance(cwGranter)
	require.Equal(t, grantedBalanceBefore.Sub(msg.Amount...), grantedBalanceAfter)
	require.Equal(t, cwGranterBalanceBefore.Sub(fees), cwGranterBalanceAfter)

	// let's now test the fallthrough case

	humanGranter := app.GetAccount(2)

	err = app.GetApp().Keepers.FeeGrantKeeper.GrantAllowance(
		app.GetContext(),
		humanGranter.Address,
		grantedAcc.Address,
		&feegrant.BasicAllowance{})
	require.NoError(t, err)

	// send a tx using fee grant
	grantedBalanceBefore = app.GetBalance(grantedAcc.Address)
	granterBalanceBefore := app.GetBalance(humanGranter.Address)
	_, _, _, err = app.SendMsgs(grantedAcc, true, []sdk.Msg{msg}, e2eTesting.WithGranter(humanGranter.Address), e2eTesting.WithMsgFees(fees))
	require.NoError(t, err)

	grantedBalanceAfter = app.GetBalance(grantedAcc.Address)
	granterBalanceAfter := app.GetBalance(humanGranter.Address)
	require.Equal(t, grantedBalanceBefore.Sub(msg.Amount...), grantedBalanceAfter)
	require.Equal(t, granterBalanceBefore.Sub(fees), granterBalanceAfter)
}
