package keeper

import (
	"cosmossdk.io/collections"
	"github.com/archway-network/archway/x/cwerrors/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// SetError stores a sudo error and queues it for deletion after a certain block height
func (k Keeper) SetSubscription(ctx sdk.Context, contractAddress sdk.AccAddress, fee sdk.Coin) (int64, error) {
	if !k.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		return -1, types.ErrContractNotFound
	}
	existingSubFound, endHeight := k.GetSubscription(ctx, contractAddress)
	if existingSubFound {
		if err := k.SubscriptionEndBlock.Remove(ctx, collections.Join(endHeight, contractAddress.Bytes())); err != nil {
			return -1, err
		}
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return -1, err
	}

	if fee.IsLT(params.SubscriptionFee) {
		return -1, types.ErrInsufficientSubscriptionFee
	}
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, contractAddress, authtypes.FeeCollectorName, sdk.NewCoins(fee))
	if err != nil {
		return -1, err
	}

	subscriptionEndHeight := ctx.BlockHeight() + params.SubscriptionPeriod
	if err = k.SubscriptionEndBlock.Set(ctx, collections.Join(subscriptionEndHeight, contractAddress.Bytes()), contractAddress.Bytes()); err != nil {
		return -1, err
	}
	return subscriptionEndHeight, k.ContractSubscriptions.Set(ctx, contractAddress, subscriptionEndHeight)
}

func (k Keeper) HasSubscription(ctx sdk.Context, contractAddress sdk.AccAddress) bool {
	has, err := k.ContractSubscriptions.Has(ctx, contractAddress)
	if err != nil {
		return false
	}
	return has
}

func (k Keeper) GetSubscription(ctx sdk.Context, contractAddress sdk.AccAddress) (bool, int64) {
	has, err := k.ContractSubscriptions.Get(ctx, contractAddress)
	if err != nil {
		return false, 0
	}
	return true, has
}

func (k Keeper) PruneSubscriptionsEndBlock(ctx sdk.Context) (err error) {
	height := ctx.BlockHeight()
	rng := collections.NewPrefixedPairRange[int64, []byte](height)
	err = k.SubscriptionEndBlock.Walk(ctx, rng, func(key collections.Pair[int64, []byte], contractAddress []byte) (bool, error) {
		if err := k.ContractSubscriptions.Remove(ctx, contractAddress); err != nil {
			return true, err
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	err = k.SubscriptionEndBlock.Clear(ctx, rng)
	return err
}
