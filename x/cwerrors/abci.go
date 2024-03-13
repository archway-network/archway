package cwerrors

import (
	"github.com/archway-network/archway/pkg"
	"github.com/archway-network/archway/x/cwerrors/keeper"
	"github.com/archway-network/archway/x/cwerrors/types"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const ErrorCallbackGasLimit = 150_000

// EndBlocker is called every block, and prunes errors that are older than the current block height.
func EndBlocker(ctx sdk.Context, k keeper.Keeper, wk types.WasmKeeperExpected) []abci.ValidatorUpdate {
	k.IterateSudoErrorCallbacks(ctx, sudoErrorCallbackExec(ctx, k, wk))
	err := k.PruneErrorsCurrentBlock(ctx)
	if err != nil {
		panic(err)
	}
	return nil
}

func sudoErrorCallbackExec(ctx sdk.Context, k keeper.Keeper, wk types.WasmKeeperExpected) func(types.SudoError) bool {
	logger := k.Logger(ctx)
	return func(sudoError types.SudoError) bool {
		sudoErrorMsg := sudoError.String()
		contractAddr := sdk.MustAccAddressFromBech32(sudoError.ContractAddress)

		logger.Debug(
			"executing error callback",
			"contract_address", sudoError.ContractAddress,
			"module_name", sudoError.ModuleName,
			"error_msg", sudoErrorMsg,
		)

		_, err := pkg.ExecuteWithGasLimit(ctx, ErrorCallbackGasLimit, func(ctx sdk.Context) error {
			_, err := wk.Sudo(ctx, contractAddr, sudoError.Bytes())
			return err
		})
		if err != nil {
			logger.Error(
				"error callback failed",
				"contract_address", sudoError.ContractAddress,
				"module_name", sudoError.ModuleName,
				"error_msg", sudoErrorMsg,
				"err", err,
			)
			err = k.StoreErrorInState(ctx, contractAddr, sudoError)
			if err != nil {
				panic(err)
			}
		}
		return false
	}
}
