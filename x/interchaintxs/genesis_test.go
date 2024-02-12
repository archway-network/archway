package interchaintxs_test

import (
	"testing"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/interchaintxs"
	"github.com/archway-network/archway/x/interchaintxs/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}
	chain := e2eTesting.NewTestChain(t, 1)
	ctx, k := chain.GetContext(), chain.GetApp().Keepers.InterchainTxsKeeper

	interchaintxs.InitGenesis(ctx, k, genesisState)
	genesis := interchaintxs.ExportGenesis(ctx, k)
	require.NotNil(t, genesis)
}
