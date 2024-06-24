package keeper_test

import (
	"testing"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/cwregistry/types"
	"github.com/stretchr/testify/require"
)

func TestSetContractMetadata(t *testing.T) {
	keeper, ctx := testutils.CWRegistryKeeper(t)
	wasmKeeper := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(wasmKeeper)

	contractAddr := e2eTesting.GenContractAddresses(1)[0]

	// TEST: No such contract exists
	err := keeper.SetContractMetadata(ctx, testutils.AccAddress(), contractAddr, types.CodeMetadata{})
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrNoSuchContract)

	// TEST: Unauthorized
	contractAdminAcc := testutils.AccAddress()
	wasmKeeper.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.String(),
	)
	wasmKeeper.AddCodeAdmin(1, contractAdminAcc.String())
	err = keeper.SetContractMetadata(ctx, testutils.AccAddress(), contractAddr, types.CodeMetadata{})
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrUnauthorized)

	// TEST: Success
	err = keeper.SetContractMetadata(ctx, contractAdminAcc, contractAddr, types.CodeMetadata{})
	require.NoError(t, err)
	require.True(t, keeper.HasCodeMetadata(ctx, 0))
}

func TestSetCodeMetadata(t *testing.T) {
	keeper, ctx := testutils.CWRegistryKeeper(t)
	wasmKeeper := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(wasmKeeper)

	// TEST: No such code exists
	err := keeper.SetCodeMetadata(ctx, testutils.AccAddress(), 1, types.CodeMetadata{})
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrNoSuchCode)

	// TEST: Unauthorized
	codeAdminAcc := testutils.AccAddress()
	wasmKeeper.AddCodeAdmin(1, codeAdminAcc.String())
	err = keeper.SetCodeMetadata(ctx, testutils.AccAddress(), 1, types.CodeMetadata{})
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrUnauthorized)

	// TEST: Success
	err = keeper.SetCodeMetadata(ctx, codeAdminAcc, 1, types.CodeMetadata{})
	require.NoError(t, err)
	require.True(t, keeper.HasCodeMetadata(ctx, 1))
}
