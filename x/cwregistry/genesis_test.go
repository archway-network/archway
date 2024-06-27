package cwregistry_test

import (
	"testing"

	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/cwregistry"
	"github.com/archway-network/archway/x/cwregistry/types"
	"github.com/stretchr/testify/require"
)

func TestExportGenesis(t *testing.T) {
	keeper, ctx := testutils.CWRegistryKeeper(t)
	wasmKeeper := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(wasmKeeper)
	senderAddr := testutils.AccAddress()
	wasmKeeper.AddCodeAdmin(1, senderAddr.String())
	wasmKeeper.AddCodeAdmin(2, senderAddr.String())
	err := keeper.SetCodeMetadata(ctx, senderAddr, 1, types.CodeMetadata{})
	require.NoError(t, err)
	err = keeper.SetCodeMetadata(ctx, senderAddr, 2, types.CodeMetadata{})
	require.NoError(t, err)

	genState := cwregistry.ExportGenesis(ctx, keeper)

	require.Equal(t, 2, len(genState.CodeMetadata))
}

func TestInitGenesis(t *testing.T) {
	keeper, ctx := testutils.CWRegistryKeeper(t)
	wasmKeeper := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(wasmKeeper)
	senderAddr := testutils.AccAddress()
	wasmKeeper.AddCodeAdmin(1, senderAddr.String())
	wasmKeeper.AddCodeAdmin(2, senderAddr.String())

	genState := types.GenesisState{
		CodeMetadata: []types.CodeMetadata{
			{
				CodeId: 1,
				Schema: "test",
			},
			{
				CodeId: 2,
				Schema: "test2",
			},
		},
	}
	cwregistry.InitGenesis(ctx, keeper, genState)

	c, err := keeper.GetAllCodeMetadata(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, len(c))
}
