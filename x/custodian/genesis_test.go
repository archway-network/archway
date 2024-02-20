package custodian_test

import (
	"testing"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/custodian"
	"github.com/archway-network/archway/x/custodian/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}
	chain := e2eTesting.NewTestChain(t, 1)
	ctx, k := chain.GetContext(), chain.GetApp().Keepers.CustodianKeeper

	custodian.InitGenesis(ctx, k, genesisState)
	genesis := custodian.ExportGenesis(ctx, k)
	require.NotNil(t, genesis)
}
