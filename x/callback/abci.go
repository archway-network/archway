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
	logger := k.Logger(ctx)
	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(err)
	}

	currentHeight := ctx.BlockHeight()
	// fetching all callbacks for current height
	callbacks, err := k.GetCallbacksByHeight(ctx, currentHeight)
	if err != nil {
		panic(err)
	}

	for _, callback := range callbacks {
		// creating CallbackMsg which is encoded to json and passed as input to contract execution
		callbackMsg := types.NewCallbackMsg(callback.GetJobId())
		// handling any panics
		defer recoverAnyPanics(logger, callback)()
		logger.Debug("contract callbacks", "contract_address", callback.ContractAddress, "job_id", callback.GetJobId())

		// creating a chiled context with limited gas meter based on configured params
		childCtx, commit := ctx.WithGasMeter(sdk.NewGasMeter(params.GetCallbackGasLimit())).CacheContext()

		// executing the callback on the contract
		if _, err := wk.Sudo(childCtx, sdk.MustAccAddressFromBech32(callback.ContractAddress), callbackMsg.Bytes()); err != nil {
			logger.Error("error executing callback", "contract_address", callback.ContractAddress, "job_id", callback.GetJobId(), "error", err)
		}

		commit()

		// deleting the callback after execution
		k.Callbacks.Remove(ctx, collections.Join3(currentHeight, sdk.MustAccAddressFromBech32(callback.ContractAddress).Bytes(), callback.GetJobId()))
	}

	return nil
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
			logger.
				Error("panic executing callback",
					"contract_address", callback.GetContractAddress(),
					"job_id", callback.GetJobId(),
					"cause", cause,
					"stacktrace", string(debug.Stack()),
				)
		}
	}
}
