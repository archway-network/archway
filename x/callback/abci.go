package callback

import (
	"fmt"
	"runtime/debug"

	"cosmossdk.io/collections"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/callback/keeper"
	"github.com/archway-network/archway/x/callback/types"
)

// EndBlocker fetches all the callbacks registered for the current block height and executes them
func EndBlocker(ctx sdk.Context, k keeper.Keeper, wk types.WasmKeeperExpected) []abci.ValidatorUpdate {
	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(err)
	}

	currentHeight := ctx.BlockHeight()
	k.IterateCallbacksByHeight(ctx, currentHeight, callbackExec(ctx, k, wk, params.GetCallbackGasLimit()))
	return nil
}

// callbackExec returns a function which executes the callback and deletes it from state after execution
func callbackExec(ctx sdk.Context, k keeper.Keeper, wk types.WasmKeeperExpected, callbackGasLimit uint64) func(types.Callback) bool {
	logger := k.Logger(ctx)
	return func(callback types.Callback) bool {
		// creating CallbackMsg which is encoded to json and passed as input to contract execution
		callbackMsg := types.NewCallbackMsg(callback.GetJobId())
		// handling any panics
		defer recoverAnyPanics(logger, callback)()
		logger.Debug(
			"executing callback",
			"contract_address", callback.ContractAddress,
			"job_id", callback.GetJobId(),
			"msg", callbackMsg.String(),
		)
		// creating a child context with limited gas meter based on configured params
		childCtx, commit := ctx.WithGasMeter(sdk.NewGasMeter(callbackGasLimit)).CacheContext()

		// executing the callback on the contract
		if _, err := wk.Sudo(childCtx, sdk.MustAccAddressFromBech32(callback.ContractAddress), callbackMsg.Bytes()); err != nil {
			logger.Error(
				"error executing callback",
				"contract_address", callback.ContractAddress,
				"job_id", callback.GetJobId(),
				"error", err,
			)
		}

		// todo: check unused gas and refund any leftover to the address which reserved the callback. will do in diff PR

		commit()

		// deleting the callback after execution
		if err := k.Callbacks.Remove(
			ctx,
			collections.Join3(
				callback.CallbackHeight,
				sdk.MustAccAddressFromBech32(callback.ContractAddress).Bytes(),
				callback.GetJobId(),
			),
		); err != nil {
			panic(err)
		}

		return false
	}
}

// recoverAnyPanics catches any panics and logs cause to error
func recoverAnyPanics(logger log.Logger, callback types.Callback) func() {
	return func() {
		if r := recover(); r != nil {
			var cause string
			switch rType := r.(type) {
			case sdk.ErrorOutOfGas:
				cause = fmt.Sprintf("out of gas in location: %v", rType.Descriptor)
			default:
				cause = fmt.Sprintf("%s", r)
			}
			logger.Error(
				"panic executing callback",
				"contract_address", callback.GetContractAddress(),
				"job_id", callback.GetJobId(),
				"cause", cause,
				"stacktrace", string(debug.Stack()),
			)
		}
	}
}
