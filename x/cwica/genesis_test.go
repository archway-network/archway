package cwica_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	cwica "github.com/archway-network/archway/x/cwica"
	"github.com/archway-network/archway/x/cwica/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}
	chain := e2eTesting.NewTestChain(t, 1)
	ctx, k := chain.GetContext(), chain.GetApp().Keepers.CWICAKeeper

	cwica.InitGenesis(ctx, k, genesisState)
	genesis := cwica.ExportGenesis(ctx, k)
	require.NotNil(t, genesis)
}
