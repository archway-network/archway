package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EmitParamsUpdatedEvent emits an event when the params are updated
func EmitParamsUpdatedEvent(ctx sdk.Context, authority string, newParams Params) {
	err := ctx.EventManager().EmitTypedEvent(&ParamsUpdatedEvent{
		Authority: authority,
		NewParams: newParams,
	})
	if err != nil {
		panic(fmt.Errorf("sending ParamsUpdatedEvent event: %w", err))
	}
}

// EmitSubscribedToErrorsEvent emits an event when a contract is subscribed to errors
func EmitSubscribedToErrorsEvent(ctx sdk.Context, sender, contractAddress string, fees sdk.Coin, subValidTill int64) {
	err := ctx.EventManager().EmitTypedEvent(&SubscribedToErrorsEvent{
		Sender:                sender,
		ContractAddress:       contractAddress,
		FeesPaid:              fees,
		SubscriptionValidTill: subValidTill,
	})
	if err != nil {
		panic(fmt.Errorf("sending SubscribedToErrorsEvent event: %w", err))
	}
}

// EmitStoringErrorEvent emits an event when an error is stored
func EmitStoringErrorEvent(ctx sdk.Context, sudoError SudoError, deletionBlockHeight int64) {
	err := ctx.EventManager().EmitTypedEvent(&StoringErrorEvent{
		Error:               sudoError,
		DeletionBlockHeight: deletionBlockHeight,
	})
	if err != nil {
		panic(fmt.Errorf("sending StoringErrorEvent event: %w", err))
	}
}

// EmitSudoErrorCallbackFailedEvent emits an event when a callback for a sudo error fails
func EmitSudoErrorCallbackFailedEvent(ctx sdk.Context, sudoError SudoError, callbackErr string) {
	err := ctx.EventManager().EmitTypedEvent(&SudoErrorCallbackFailedEvent{
		Error:                sudoError,
		CallbackErrorMessage: callbackErr,
	})
	if err != nil {
		panic(fmt.Errorf("sending SudoErrorCallbackFailedEvent event: %w", err))
	}
}
