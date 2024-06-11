package cwfees_test

import (
	"encoding/json"
	"fmt"
	"testing"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/cwfees/types"
)

func TestKeeper(t *testing.T) {
	app := e2eTesting.NewTestChain(t, 0)
	k := app.GetApp().Keepers.CWFeesKeeper
	ctx := app.GetContext()

	t.Run("register as granter – not a cw contract", func(t *testing.T) {
		acc := app.GetAccount(0)
		err := k.RegisterAsGranter(ctx, acc.Address)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrNotAContract)
	})

	t.Run("unregister as granter – not a granter", func(t *testing.T) {
		acc := app.GetAccount(0)
		err := k.UnregisterAsGranter(ctx, acc.Address)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrNotAGranter)
	})

	t.Run("register as granter OK – unregister as granter OK", func(t *testing.T) {
		codeID := app.UploadContract(app.GetAccount(0), "../../contracts/cwfees/artifacts/cwfees.wasm", wasmdTypes.DefaultUploadAccess)
		grantedAcc := app.GetAccount(1) // account who receives grants.
		initMsg := fmt.Sprintf(`{"grants": ["%s"]}`, grantedAcc.Address)
		cwGranter, _ := app.InstantiateContract(app.GetAccount(0), codeID, app.GetAccount(0).Address.String(), "cwfees", sdk.NewCoins(sdk.NewInt64Coin("stake", 1_000_000_000_000)), json.RawMessage(initMsg))
		ctx := app.GetContext()
		err := k.RegisterAsGranter(ctx, cwGranter)
		require.NoError(t, err)
		isGranter, err := k.IsGrantingContract(ctx, cwGranter)
		require.NoError(t, err)
		require.True(t, isGranter)

		err = k.UnregisterAsGranter(ctx, cwGranter)
		require.NoError(t, err)
		isGranter, err = k.IsGrantingContract(ctx, cwGranter)
		require.NoError(t, err)
		require.False(t, isGranter)
	})

	t.Run("state import and export", func(t *testing.T) {
		codeID := app.UploadContract(app.GetAccount(0), "../../contracts/cwfees/artifacts/cwfees.wasm", wasmdTypes.DefaultUploadAccess)
		grantedAcc := app.GetAccount(1) // account who receives grants.
		initMsg := fmt.Sprintf(`{"grants": ["%s"]}`, grantedAcc.Address)
		cwGranter, _ := app.InstantiateContract(app.GetAccount(0), codeID, app.GetAccount(0).Address.String(), "cwfees", sdk.NewCoins(sdk.NewInt64Coin("stake", 1_000_000_000_000)), json.RawMessage(initMsg))

		ctx := app.GetContext()
		wantState := &types.GenesisState{GrantingContracts: []string{cwGranter.String()}}
		err := k.ImportState(ctx, wantState)
		require.NoError(t, err)

		gotState, err := k.ExportState(ctx)
		require.NoError(t, err)

		require.Equal(t, wantState, gotState)
	})
}

// func TestFullIntegration(t *testing.T) {
// 	app := e2eTesting.NewTestChain(t, 0, e2eTesting.WithGenAccounts(10))
// 	deployer := app.GetAccount(0)

// 	codeID := app.UploadContract(deployer, "../../contracts/cwfees/artifacts/cwfees.wasm", wasmdTypes.DefaultUploadAccess)

// 	grantedAcc := app.GetAccount(1) // account who receives grants.
// 	initMsg := fmt.Sprintf(`{"grants": ["%s"]}`, grantedAcc.Address)
// 	cwGranter, _ := app.InstantiateContract(deployer, codeID, deployer.Address.String(), "cwfees", sdk.NewCoins(sdk.NewInt64Coin("stake", 1_000_000_000_000)), json.RawMessage(initMsg))

// 	// register as cwfees contract.
// 	err := app.GetApp().Keepers.CWFeesKeeper.RegisterAsGranter(app.GetContext(), cwGranter)
// 	require.NoError(t, err)

// 	// now try to send a tx with a cw granter
// 	msg := &banktypes.MsgSend{
// 		FromAddress: grantedAcc.Address.String(),
// 		ToAddress:   deployer.Address.String(),
// 		Amount:      sdk.NewCoins(sdk.NewInt64Coin("stake", 1)),
// 	}

// 	grantedBalanceBefore := app.GetBalance(grantedAcc.Address)
// 	cwGranterBalanceBefore := app.GetBalance(cwGranter)
// 	fees := sdk.NewInt64Coin("stake", 100_000_000_000)
// 	_, _, _, err = app.SendMsgs(grantedAcc, true, []sdk.Msg{msg}, e2eTesting.WithGranter(cwGranter), e2eTesting.WithMsgFees(fees))
// 	require.NoError(t, err)

// 	grantedBalanceAfter := app.GetBalance(grantedAcc.Address)
// 	cwGranterBalanceAfter := app.GetBalance(cwGranter)
// 	require.Equal(t, grantedBalanceBefore.Sub(msg.Amount...), grantedBalanceAfter)
// 	require.Equal(t, cwGranterBalanceBefore.Sub(fees), cwGranterBalanceAfter)

// 	// let's now test the fallthrough case

// 	humanGranter := app.GetAccount(2)

// 	err = app.GetApp().Keepers.FeeGrantKeeper.GrantAllowance(
// 		app.GetContext(),
// 		humanGranter.Address,
// 		grantedAcc.Address,
// 		&feegrant.BasicAllowance{})
// 	require.NoError(t, err)

// 	// send a tx using fee grant
// 	grantedBalanceBefore = app.GetBalance(grantedAcc.Address)
// 	granterBalanceBefore := app.GetBalance(humanGranter.Address)
// 	_, _, _, err = app.SendMsgs(grantedAcc, true, []sdk.Msg{msg}, e2eTesting.WithGranter(humanGranter.Address), e2eTesting.WithMsgFees(fees))
// 	require.NoError(t, err)

// 	grantedBalanceAfter = app.GetBalance(grantedAcc.Address)
// 	granterBalanceAfter := app.GetBalance(humanGranter.Address)
// 	require.Equal(t, grantedBalanceBefore.Sub(msg.Amount...), grantedBalanceAfter)
// 	require.Equal(t, granterBalanceBefore.Sub(fees), granterBalanceAfter)

// 	// send a malicious tx where the contract is spending all the gas in ante handler
// 	// computation. Adding malicious as denom triggers the endless loop on the contract.
// 	// We expect this loop to finish on RequestGrantGasLimit.
// 	txGasLimit := uint64(1_000_000)
// 	maliciousFee := sdk.NewInt64Coin("malicious", 1)
// 	gasInfo, _, _, err := app.SendMsgs(grantedAcc, false, []sdk.Msg{msg}, e2eTesting.WithGranter(cwGranter), e2eTesting.WithMsgFees(sdk.NewCoins(fees, maliciousFee)...), e2eTesting.WithTxGasLimit(txGasLimit))
// 	require.Error(t, err)
// 	// we expect the contract to have consumed less than the tx gas limit,
// 	// as RequestGrantGasLimit will have shutdown the execution before.
// 	require.Less(t, gasInfo.GasUsed, txGasLimit)
// }
