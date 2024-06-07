package interchaintest

import (
	"context"
	"testing"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
)

func TestChainStart(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	archwayChain, _, ctx := startChain(t, "local")

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*45)
	defer timeoutCtxCancel()

	// just wait for 10 blocks to be produced
	err := testutil.WaitForBlocks(timeoutCtx, 10, archwayChain)
	require.NoError(t, err, "chain did not produce blocks")
}
