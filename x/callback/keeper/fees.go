package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// EstimateCallbackFees returns the fees that will be charged for registering a callback at the given block height
func (k Keeper) EstimateCallbackFees(ctx sdk.Context, blockHeight int64) (sdk.Coin, sdk.Coin, sdk.Coin, error) {
	if blockHeight < ctx.BlockHeight() {
		return sdk.DecCoin{}, sdk.DecCoin{}, sdk.DecCoin{}, status.Errorf(codes.InvalidArgument, "block height %d is in the past", blockHeight)
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return sdk.DecCoin{}, sdk.DecCoin{}, sdk.DecCoin{}, status.Errorf(codes.NotFound, "could not fetch the module params: %s", err.Error())
	}

	// Calculates the fees based on how far in the future the callback is registered
	futureReservationThreshold := ctx.BlockHeight() + int64(params.MaxFutureReservationLimit)
	if blockHeight > futureReservationThreshold {
		return sdk.DecCoin{}, sdk.DecCoin{}, sdk.DecCoin{}, status.Errorf(codes.OutOfRange, "block height %d is too far in the future. max block height callback can be registered at %d", blockHeight, futureReservationThreshold)
	}
	// futureReservationFeeMultiplies * (requestBlockHeight - currentBlockHeight)
	futureReservationFeesAmount := params.FutureReservationFeeMultiplier.MulInt64((blockHeight - ctx.BlockHeight()))
	futureReservationFee := sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, futureReservationFeesAmount)

	// Calculates the fees based on how many callbacks are registered at the given block height
	callbacksForHeight, err := k.GetCallbacksByHeight(ctx, blockHeight)
	if err != nil {
		return sdk.DecCoin{}, sdk.DecCoin{}, sdk.DecCoin{}, status.Errorf(codes.NotFound, "could not fetch callbacks for given height: %s", err.Error())
	}
	totalCallbacks := len(callbacksForHeight)
	if totalCallbacks >= int(params.MaxBlockReservationLimit) {
		return sdk.DecCoin{}, sdk.DecCoin{}, sdk.DecCoin{}, status.Errorf(codes.OutOfRange, "block height %d has reached max reservation limit", blockHeight)
	}
	// blockReservatuiionFeeMultiplier * totalCallbacksRegistered
	blockReservationFeesAmount := params.BlockReservationFeeMultiplier.MulInt64(int64(totalCallbacks))
	blockReservationFee := sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, blockReservationFeesAmount)

	// Calculates the fees based on the max gas limit of the callback and current price of gas
	transactionFeeAmount := k.rewardsKeeper.ComputationalPriceOfGas(ctx).Amount.MulInt64(int64(params.GetCallbackGasLimit()))
	transactionFee := sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, transactionFeeAmount)
	return futureReservationFee, blockReservationFee, transactionFee, nil
}
