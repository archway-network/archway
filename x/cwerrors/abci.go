package cwerrors

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/pkg"
	"github.com/archway-network/archway/x/cwerrors/keeper"
	"github.com/archway-network/archway/x/cwerrors/types"
)

const ErrorCallbackGasLimit = 150_000

// EndBlocker is called every block, and prunes errors that are older than the current block height.
func EndBlocker(ctx sdk.Context, k keeper.Keeper, wk types.WasmKeeperExpected) []abci.ValidatorUpdate {
	// Iterate over all errors (with callback subscription) and execute the error callback for each error
	k.IterateSudoErrorCallbacks(ctx, sudoErrorCallbackExec(ctx, k, wk))
	// Prune any error callback subscripitons that have expired in the current block height
	if err := k.PruneSubscriptionsEndBlock(ctx); err != nil {
		panic(err)
	}
	// Prune any errors(in state) that have expired in the current block height
	if err := k.PruneErrorsCurrentBlock(ctx); err != nil {
		panic(err)
	}

	return nil
}

func sudoErrorCallbackExec(ctx sdk.Context, k keeper.Keeper, wk types.WasmKeeperExpected) func(types.SudoError) bool {
	return func(sudoError types.SudoError) bool {
		contractAddr := sdk.MustAccAddressFromBech32(sudoError.ContractAddress)

		sudoMsg := types.NewSudoMsg(sudoError)
		_, err := pkg.ExecuteWithGasLimit(ctx, ErrorCallbackGasLimit, func(ctx sdk.Context) error {
			_, err := wk.Sudo(ctx, contractAddr, sudoMsg.Bytes())
			return err
		})
		if err != nil {
			// In case Sudo error, such as out of gas, emit an event and store the error in state (so that the error is not lost)
			types.EmitSudoErrorCallbackFailedEvent(
				ctx,
				sudoError,
				err.Error(),
			)
			newSudoErr := types.SudoError{
				ModuleName:      types.ModuleName,
				ContractAddress: sudoError.ContractAddress,
				ErrorCode:       int32(1),
				InputPayload:    string(sudoError.Bytes()),
				ErrorMessage:    err.Error() + "---" + sudoMsg.String(),
			}
			err = k.StoreErrorInState(ctx, contractAddr, newSudoErr)
			if err != nil {
				panic(err)
			}
		}
		return false
	}
}
