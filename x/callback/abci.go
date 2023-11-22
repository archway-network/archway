package callback

import (
	"fmt"
	"runtime/debug"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/callback/keeper"
	"github.com/archway-network/archway/x/callback/types"
)

// EndBlocker fetches all the callbacks registered for the current block height and executes them
func EndBlocker(ctx sdk.Context, k keeper.Keeper, wk types.WasmKeeperExpected) []abci.ValidatorUpdate {
	logger := k.Logger(ctx)
	currentHeight := ctx.BlockHeight()
	callbacks, err := k.GetCallbacksByHeight(ctx, currentHeight)
	if err != nil {
		panic(err)
	}
	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(err)
	}

	for _, callback := range callbacks {
		callbackMsg := types.NewCallbackMsg(callback.GetJobId())
		defer recoverAnyPanics(logger, callback)()
		logger.Debug("contract callbacks", "contract_address", callback.ContractAddress, "job_id", callback.GetJobId())

		childCtx, commit := ctx.WithGasMeter(sdk.NewGasMeter(params.GetCallbackGasLimit())).CacheContext()

		if _, err := wk.Sudo(childCtx, sdk.MustAccAddressFromBech32(callback.ContractAddress), callbackMsg.Bytes()); err != nil {
			logger.Error("error executing callback", "contract_address", callback.ContractAddress, "job_id", callback.GetJobId(), "error", err)
		}

		ctx.EventManager().EmitEvents(childCtx.EventManager().Events())
		commit()
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
					"contract-address", callback.GetContractAddress(),
					"job_id", callback.GetJobId(),
					"cause", cause,
					"stacktrace", string(debug.Stack()),
				)
		}
	}
}
