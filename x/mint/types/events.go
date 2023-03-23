package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EmitBlockInflationEvent(ctx sdk.Context, tokens sdk.Dec, inflation sdk.Dec) {
	err := ctx.EventManager().EmitTypedEvent(&BlockInflationEvent{
		MintAmount: tokens,
		Inflation:  inflation,
	})
	if err != nil {
		panic(fmt.Errorf("sending BlockInflationEvent event: %w", err))
	}
}
