package cwerrors_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/cwerrors"
	"github.com/archway-network/archway/x/cwerrors/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestExportGenesis(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	ctx, keeper := chain.GetContext(), chain.GetApp().Keepers.CWErrorsKeeper

	exportedState := cwerrors.ExportGenesis(ctx, keeper)
	require.Equal(t, types.DefaultParams(), exportedState.Params)

	newParams := types.Params{
		ErrorStoredTime:       99999,
		DisableErrorCallbacks: true,
		SubscriptionFee:       sdk.NewCoin("stake", sdk.NewInt(100)),
		SubscriptionPeriod:    1,
	}
	err := keeper.SetParams(ctx, newParams)
	require.NoError(t, err)

	exportedState = cwerrors.ExportGenesis(ctx, keeper)
	require.Equal(t, newParams.DisableErrorCallbacks, exportedState.Params.DisableErrorCallbacks)
	require.Equal(t, newParams.ErrorStoredTime, exportedState.Params.ErrorStoredTime)
	require.Equal(t, newParams.SubscriptionFee, exportedState.Params.SubscriptionFee)
	require.Equal(t, newParams.SubscriptionPeriod, exportedState.Params.SubscriptionPeriod)
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
			SubscriptionFee:       sdk.NewCoin("stake", sdk.NewInt(100)),
			SubscriptionPeriod:    1,
		},
	}
	cwerrors.InitGenesis(ctx, keeper, genstate)

	params, err = keeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, genstate.Params.DisableErrorCallbacks, params.DisableErrorCallbacks)
	require.Equal(t, genstate.Params.ErrorStoredTime, params.ErrorStoredTime)
	require.Equal(t, genstate.Params.SubscriptionFee, params.SubscriptionFee)
	require.Equal(t, genstate.Params.SubscriptionPeriod, params.SubscriptionPeriod)
}
