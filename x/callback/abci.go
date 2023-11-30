package callback

import (
	"cosmossdk.io/collections"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/pkg"
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

		logger.Debug(
			"executing callback",
			"contract_address", callback.ContractAddress,
			"job_id", callback.GetJobId(),
			"msg", callbackMsg.String(),
		)

		gasUsed, err := pkg.ExecuteWithGasLimit(ctx, callbackGasLimit, func(ctx sdk.Context) error {
			// executing the callback on the contract
			_, err := wk.Sudo(ctx, sdk.MustAccAddressFromBech32(callback.ContractAddress), callbackMsg.Bytes())
			return err
		})
		if err != nil {
			logger.Error(
				"error executing callback",
				"contract_address", callback.ContractAddress,
				"job_id", callback.GetJobId(),
				"error", err,
			)
			// todo: throw error event with details on failure. will do in diff PR
		}

		unusedGas := callbackGasLimit - gasUsed
		logger.Info(
			"callback executed with pending gas",
			"contract_address", callback.ContractAddress,
			"job_id", callback.GetJobId(),
			"unused_gas", unusedGas,
		)
		// todo: refund any leftover to the address which reserved the callback. will do in diff PR

		// deleting the callback after execution
		if err := k.Callbacks.Remove(
			ctx,
			collections.Join3(
				callback.CallbackHeight,
				sdk.MustAccAddressFromBech32(callback.ContractAddress).Bytes(),
				callback.GetJobId(),
			),
		); err != nil {
			logger.Error(
				"error deleting callback",
				"contract_address", callback.ContractAddress,
				"job_id", callback.GetJobId(),
				"error", err,
			)
			// todo: throw error event with details on failure. will do in diff PR
		}

		return false
	}
}
