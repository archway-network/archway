package cwgrant_test

import (
	"testing"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/cwgrant/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
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

}
