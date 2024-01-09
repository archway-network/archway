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
			// Emit failure event
			types.EmitCallbackExecutedFailedEvent(
				ctx,
				callback.ContractAddress,
				callback.GetJobId(),
				callbackMsg.String(),
				gasUsed,
				err.Error(),
			)

			// This is because gasUsed amount returned is greater than the gas limit. cuz ofc.
			// so we set it to callbackGasLimit so when we do txFee refund, we arent trying to refund more than we should
			// e.g if callbackGasLimit is 10, but gasUsed is 100, we need to use 10 to calculate txFeeRefund.
			// else the module will pay back more than it took from the user 💀
			// TLDR; this ensures in case of "out of gas error", we keep all txFees and refund nothing.
			gasUsed = callbackGasLimit
		} else {
			logger.Info(
				"callback executed successfully",
				"contract_address", callback.ContractAddress,
				"job_id", callback.GetJobId(),
				"msg", callbackMsg.String(),
				"gas_used", gasUsed,
			)
			// Emit success event
			types.EmitCallbackExecutedSuccessEvent(
				ctx,
				callback.ContractAddress,
				callback.GetJobId(),
				callbackMsg.String(),
				gasUsed,
			)
		}

		logger.Info(
			"callback executed with pending gas",
			"contract_address", callback.ContractAddress,
			"job_id", callback.GetJobId(),
			"used_gas", gasUsed,
		)

		// Calculate current tx fees based on gasConsumed. Refund any leftover to the address which reserved the callback
		txFeesConsumed := k.CalculateTransactionFees(ctx, gasUsed)
		if txFeesConsumed.IsLT(*callback.FeeSplit.TransactionFees) {
			refundAmount := callback.FeeSplit.TransactionFees.Sub(txFeesConsumed)
			err := k.RefundFromCallbackModule(ctx, callback.ReservedBy, refundAmount)
			if err != nil {
				panic(err)
			}
		}

		// Send fees to fee collector
		feeCollectorAmount := callback.FeeSplit.BlockReservationFees.
			Add(*callback.FeeSplit.FutureReservationFees).
			Add(*callback.FeeSplit.SurplusFees).
			Add(txFeesConsumed)
		err = k.SendToFeeCollector(ctx, feeCollectorAmount)
		if err != nil {
			panic(err)
		}

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
