package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EmitBlockInflationEvent emits the BlockInflationEvent event
func EmitBlockInflationEvent(ctx sdk.Context, tokens sdk.Dec, inflation sdk.Dec) {
	err := ctx.EventManager().EmitTypedEvent(&BlockInflationEvent{
		MintAmount: tokens,
		Inflation:  inflation,
	})
	if err != nil {
		panic(fmt.Errorf("sending BlockInflationEvent event: %w", err))
	}
}

// EmitBlockInflationDistributionEvent emits the BlockInflationDistributionEvent event
func EmitBlockInflationDistributionEvent(ctx sdk.Context, recipient string, tokens sdk.Coin) {
	err := ctx.EventManager().EmitTypedEvent(&BlockInflationDistributionEvent{
		Recipient: recipient,
		Tokens:    tokens,
	})
	if err != nil {
		panic(fmt.Errorf("sending BlockInflationDistributionEvent event: %w", err))
	}
}
