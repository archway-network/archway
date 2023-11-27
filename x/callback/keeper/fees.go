package keeper

import (
	"github.com/archway-network/archway/x/callback/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) EstimateCallbackFees(request *types.QueryEstimateCallbackFeesRequest, ctx sdk.Context) (sdk.DecCoin, sdk.DecCoin, sdk.DecCoin, error) {
	if request.BlockHeight < ctx.BlockHeight() {
		return sdk.DecCoin{}, sdk.DecCoin{}, sdk.DecCoin{}, status.Errorf(codes.InvalidArgument, "block height %d is in the past", request.BlockHeight)
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return sdk.DecCoin{}, sdk.DecCoin{}, sdk.DecCoin{}, status.Errorf(codes.NotFound, "could not fetch the module params: %s", err.Error())
	}

	// Calculates the fees based on how far in the future the callback is registered
	futureReservationThreshold := ctx.BlockHeight() + int64(params.MaxFutureReservationLimit)
	if request.BlockHeight > futureReservationThreshold {
		return sdk.DecCoin{}, sdk.DecCoin{}, sdk.DecCoin{}, status.Errorf(codes.OutOfRange, "block height %d is too far in the future. max block height callback can be registered at %d", request.BlockHeight, futureReservationThreshold)
	}
	// futureReservationFeeMultiplies * (requestBlockHeight - currentBlockHeight)
	futureReservationFeesAmount := params.FutureReservationFeeMultiplier.MulInt64((request.GetBlockHeight() - ctx.BlockHeight()))
	futureReservationFee := sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, futureReservationFeesAmount)

	// Calculates the fees based on how many callbacks are registered at the given block height
	callbacksForHeight, err := k.GetCallbacksByHeight(ctx, request.GetBlockHeight())
	if err != nil {
		return sdk.DecCoin{}, sdk.DecCoin{}, sdk.DecCoin{}, status.Errorf(codes.NotFound, "could not fetch callbacks for given height: %s", err.Error())
	}
	totalCallbacks := len(callbacksForHeight)
	if totalCallbacks >= int(params.MaxBlockReservationLimit) {
		return sdk.DecCoin{}, sdk.DecCoin{}, sdk.DecCoin{}, status.Errorf(codes.OutOfRange, "block height %d has reached max reservation limit", request.BlockHeight)
	}
	// blockReservatuiionFeeMultiplier * totalCallbacksRegistered
	blockReservationFeesAmount := params.BlockReservationFeeMultiplier.MulInt64(int64(totalCallbacks))
	blockReservationFee := sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, blockReservationFeesAmount)

	// Calculates the fees based on the max gas limit of the callback and current price of gas
	transactionFeeAmount := k.rewardsKeeper.ComputationalPriceOfGas(ctx).Amount.MulInt64(int64(params.GetCallbackGasLimit()))
	transactionFee := sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, transactionFeeAmount)
	return futureReservationFee, blockReservationFee, transactionFee, nil
}
