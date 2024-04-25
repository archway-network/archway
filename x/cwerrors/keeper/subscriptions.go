package keeper

import (
	"strings"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/archway-network/archway/x/cwerrors/types"
)

// SetSubscription sets a subscription for a contract so the contract can receive error callbacks
func (k Keeper) SetSubscription(ctx sdk.Context, sender, contractAddress sdk.AccAddress, fee sdk.Coin) (int64, error) {
	if !k.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		return -1, types.ErrContractNotFound
	}
	if !isAuthorizedToSubscribe(ctx, k, contractAddress, sender.String()) {
		return -1, types.ErrUnauthorized
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

	if !fee.IsEqual(params.SubscriptionFee) {
		return -1, types.ErrIncorrectSubscriptionFee
	}
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, authtypes.FeeCollectorName, sdk.NewCoins(fee))
	if err != nil {
		return -1, err
	}

	subscriptionEndHeight := ctx.BlockHeight() + params.SubscriptionPeriod
	if err = k.SubscriptionEndBlock.Set(ctx, collections.Join(subscriptionEndHeight, contractAddress.Bytes()), contractAddress.Bytes()); err != nil {
		return -1, err
	}
	return subscriptionEndHeight, k.ContractSubscriptions.Set(ctx, contractAddress, subscriptionEndHeight)
}

// HasSubscription checks if a contract has a subscription
func (k Keeper) HasSubscription(ctx sdk.Context, contractAddress sdk.AccAddress) bool {
	has, err := k.ContractSubscriptions.Has(ctx, contractAddress)
	if err != nil {
		return false
	}
	return has
}

// GetSubscription returns the subscription end height for a contract
func (k Keeper) GetSubscription(ctx sdk.Context, contractAddress sdk.AccAddress) (bool, int64) {
	has, err := k.ContractSubscriptions.Get(ctx, contractAddress)
	if err != nil {
		return false, 0
	}
	return true, has
}

// PruneSubscriptionsEndBlock prunes subscriptions that have ended at the given block. This is executed at the module endblocker
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
	return k.SubscriptionEndBlock.Clear(ctx, rng)
}

// isAuthorizedToSubscribe checks if the sender is authorized to subscribe to the contract
func isAuthorizedToSubscribe(ctx sdk.Context, k Keeper, contractAddress sdk.AccAddress, sender string) bool {
	if strings.EqualFold(sender, contractAddress.String()) { // A contract can set subscriptions for itself
		return true
	}

	contractInfo := k.wasmKeeper.GetContractInfo(ctx, contractAddress)
	if strings.EqualFold(sender, contractInfo.Admin) { // Admin of the contract can set subscriptions
		return true
	}

	contractMetadata := k.rewardsKeeper.GetContractMetadata(ctx, contractAddress)
	return contractMetadata != nil && strings.EqualFold(sender, contractMetadata.OwnerAddress) // Owner of the contract can set subscriptions
}
