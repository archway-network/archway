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

	archwayChain, client, ctx := startChain(t, "local")

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*45)
	defer timeoutCtxCancel()

	// just wait for 10 blocks to be produced
	err := testutil.WaitForBlocks(timeoutCtx, 10, archwayChain)
	require.NoError(t, err, "chain did not produce blocks")

	t.Cleanup(func() {
		err = client.Close()
		if err != nil {
			t.Logf("an error occurred while closing the chain: %s", err)
		}
	})
}
