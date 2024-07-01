package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/cwregistry/types"
)

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
