package cwerrors_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/cwerrors"
	"github.com/archway-network/archway/x/cwerrors/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestExportGenesis(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	ctx, keeper := chain.GetContext(), chain.GetApp().Keepers.CWErrorsKeeper
	contractViewer := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	contractAddresses := e2eTesting.GenContractAddresses(3)
	contractAddr := contractAddresses[0]
	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAddr.String(),
	)
	err := keeper.SetError(ctx, types.SudoError{ContractAddress: contractAddr.String(), ModuleName: "test"})
	require.NoError(t, err)
	err = keeper.SetError(ctx, types.SudoError{ContractAddress: contractAddr.String(), ModuleName: "test"})
	require.NoError(t, err)

	exportedState := cwerrors.ExportGenesis(ctx, keeper)
	require.Equal(t, types.DefaultParams(), exportedState.Params)
	require.Len(t, exportedState.Errors, 2)

	newParams := types.Params{
		ErrorStoredTime:    99999,
		SubscriptionFee:    sdk.NewCoin("stake", sdk.NewInt(100)),
		SubscriptionPeriod: 1,
	}
	err = keeper.SetParams(ctx, newParams)
	require.NoError(t, err)

	exportedState = cwerrors.ExportGenesis(ctx, keeper)
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
			ErrorStoredTime:    99999,
			SubscriptionFee:    sdk.NewCoin("stake", sdk.NewInt(100)),
			SubscriptionPeriod: 1,
		},
		Errors: []types.SudoError{
			{ContractAddress: "addr1", ModuleName: "test"},
		},
	}
	cwerrors.InitGenesis(ctx, keeper, genstate)

	params, err = keeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, genstate.Params.ErrorStoredTime, params.ErrorStoredTime)
	require.Equal(t, genstate.Params.SubscriptionFee, params.SubscriptionFee)
	require.Equal(t, genstate.Params.SubscriptionPeriod, params.SubscriptionPeriod)

	sudoErrs, err := keeper.ExportErrors(ctx)
	require.NoError(t, err)
	require.Len(t, sudoErrs, 0)
}
