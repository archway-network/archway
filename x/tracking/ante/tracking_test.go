package ante_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/tracking/ante"
)

func TestTrackingAnteHandler(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	ctx, keeper := chain.GetContext(), chain.GetApp().TrackingKeeper

	anteHandler := ante.NewTxGasTrackingDecorator(keeper)
	for i := uint64(1); i < 10; i++ {
		_, err := anteHandler.AnteHandle(ctx, nil, false, testutils.NoopAnteHandler)
		require.NoError(t, err)

		assert.Equal(t, i, keeper.GetCurrentTxID(ctx))
	}
}
