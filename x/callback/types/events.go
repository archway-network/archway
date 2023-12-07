package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EmitCallbackRegisteredEvent(
	ctx sdk.Context,
	contractAddress string,
	jobId uint64,
	callbackHeight int64,
	feeSplit *CallbackFeesFeeSplit,
	reservedBy string,
) {
	err := ctx.EventManager().EmitTypedEvent(&CallbackRegisteredEvent{
		ContractAddress: contractAddress,
		JobId:           jobId,
		CallbackHeight:  callbackHeight,
		FeeSplit:        feeSplit,
		ReservedBy:      reservedBy,
	})
	if err != nil {
		panic(fmt.Errorf("sending CallbackRegisteredEvent event: %w", err))
	}
}

func EmitCallbackCancelledEvent(
	ctx sdk.Context,
	contractAddress string,
	jobId uint64,
	callbackHeight int64,
	cancelledBy string,
	refundAmount sdk.Coin,
) {
	err := ctx.EventManager().EmitTypedEvent(&CallbackCancelledEvent{
		ContractAddress: contractAddress,
		JobId:           jobId,
		CallbackHeight:  callbackHeight,
		CancelledBy:     cancelledBy,
		RefundAmount:    refundAmount,
	})
	if err != nil {
		panic(fmt.Errorf("sending CallbackCancelledEvent event: %w", err))
	}
}

func EmitCallbackExecutedSuccessEvent(
	ctx sdk.Context,
	contractAddress string,
	jobId uint64,
	sudoMsg string,
	gasUsed uint64,
) {
	execution := NewCallbackExecuted(contractAddress, jobId, sudoMsg, gasUsed)
	err := ctx.EventManager().EmitTypedEvent(&CallbackExecutedSuccessEvent{
		Execution: &execution,
	})
	if err != nil {
		panic(fmt.Errorf("sending CallbackExecutedSuccessEvent event: %w", err))
	}
}

func EmitCallbackExecutedFailedEvent(
	ctx sdk.Context,
	contractAddress string,
	jobId uint64,
	sudoMsg string,
	gasUsed uint64,
	errMsg string,
) {
	execution := NewCallbackExecuted(contractAddress, jobId, sudoMsg, gasUsed)
	err := ctx.EventManager().EmitTypedEvent(&CallbackExecutedFailedEvent{
		Error:     errMsg,
		Execution: &execution,
	})
	if err != nil {
		panic(fmt.Errorf("sending CallbackExecutedFailedEvent event: %w", err))
	}
}

func NewCallbackExecuted(
	contractAddress string,
	jobId uint64,
	sudoMsg string,
	gasUsed uint64,
) CallbackExecuted {
	execution := CallbackExecuted{
		ContractAddress: contractAddress,
		JobId:           jobId,
		SudoMsg:         sudoMsg,
		GasUsed:         gasUsed,
	}
	return execution
}
