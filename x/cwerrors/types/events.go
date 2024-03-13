package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EmitParamsUpdatedEvent(ctx sdk.Context, authority string, newParams Params) {
	err := ctx.EventManager().EmitTypedEvent(&ParamsUpdatedEvent{
		Authority: authority,
		NewParams: newParams,
	})
	if err != nil {
		panic(fmt.Errorf("sending ParamsUpdatedEvent event: %w", err))
	}
}

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

func EmitStoringErrorEvent(ctx sdk.Context, sudoError SudoError, deletionBlockHeight int64) {
	err := ctx.EventManager().EmitTypedEvent(&StoringErrorEvent{
		Error:               sudoError,
		DeletionBlockHeight: deletionBlockHeight,
	})
	if err != nil {
		panic(fmt.Errorf("sending StoringErrorEvent event: %w", err))
	}
}

func EmitSudoErrorCallbackFailedEvent(ctx sdk.Context, sudoError SudoError, callbackErr string) {
	err := ctx.EventManager().EmitTypedEvent(&SudoErrorCallbackFailedEvent{
		Error:                sudoError,
		CallbackErrorMessage: callbackErr,
	})
	if err != nil {
		panic(fmt.Errorf("sending SudoErrorCallbackFailedEvent event: %w", err))
	}
}
