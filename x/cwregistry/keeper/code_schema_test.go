package keeper_test

import (
	"testing"

	"github.com/archway-network/archway/pkg/testutils"
	"github.com/stretchr/testify/require"
)

func TestSetSchema(t *testing.T) {
	keeper, ctx := testutils.CWRegistryKeeper(t)

	// Getting schema which doesn't exist
	_, err := keeper.GetSchema(ctx, 1)
	require.Error(t, err)

	// Saving schema
	err = keeper.SetSchema(ctx, 1, "testContent")
	require.NoError(t, err)

	// Getting schema which now exists
	schema, err := keeper.GetSchema(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, "testContent", schema)
}
