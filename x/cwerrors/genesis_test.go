package cwerrors_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/cwerrors"
	"github.com/archway-network/archway/x/cwerrors/types"
)

func TestExportGenesis(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	ctx, keeper := chain.GetContext(), chain.GetApp().Keepers.CWErrorsKeeper

	exportedState := cwerrors.ExportGenesis(ctx, keeper)
	require.Equal(t, types.DefaultParams(), exportedState.Params)

	newParams := types.Params{
		ErrorStoredTime:       99999,
		DisableErrorCallbacks: true,
	}
	err := keeper.SetParams(ctx, newParams)
	require.NoError(t, err)

	exportedState = cwerrors.ExportGenesis(ctx, keeper)
	require.Equal(t, newParams, exportedState.Params)
}

func TestInitGenesis(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	ctx, keeper := chain.GetContext(), chain.GetApp().Keepers.CWErrorsKeeper

	genstate := types.GenesisState{
		Params: types.DefaultGenesis().Params,
	}
	cwerrors.InitGenesis(ctx, keeper, genstate)

	params, err := keeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams(), params)

	genstate = types.GenesisState{
		Params: types.Params{
			ErrorStoredTime:       99999,
			DisableErrorCallbacks: true,
		},
	}
	cwerrors.InitGenesis(ctx, keeper, genstate)

	params, err = keeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, genstate.Params, params)
}
