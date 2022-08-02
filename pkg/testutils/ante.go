package testutils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NoopAnteHandler implements the no-op AnteHandler.
func NoopAnteHandler(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
	return ctx, nil
}
