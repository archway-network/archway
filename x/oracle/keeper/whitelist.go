package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/types/set"
	"github.com/archway-network/archway/x/oracle/asset"
)

// IsWhitelistedPair returns existence of a pair in the voting target list
func (k Keeper) IsWhitelistedPair(ctx sdk.Context, pair asset.Pair) (bool, error) {
	return k.WhitelistedPairs.Has(ctx, pair)
}

// GetWhitelistedPairs returns the whitelisted pairs list on current vote period
func (k Keeper) GetWhitelistedPairs(ctx sdk.Context) ([]asset.Pair, error) {
	iter, err := k.WhitelistedPairs.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	keys, err := iter.Keys()
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// RefreshWhitelist updates the whitelist by detecting possible changes between
// the current vote targets and the current updated whitelist.
func (k Keeper) RefreshWhitelist(ctx sdk.Context, nextWhitelist []asset.Pair, currentWhitelist set.Set[asset.Pair]) {
	updateRequired := false

	if currentWhitelist.Len() != len(nextWhitelist) {
		updateRequired = true
	} else {
		for _, pair := range nextWhitelist {
			_, exists := currentWhitelist[pair]
			if !exists {
				updateRequired = true
				break
			}
		}
	}

	if updateRequired {
		err := k.WhitelistedPairs.Clear(ctx, nil)
		if err != nil {
			panic(err)
		}
		for _, pair := range nextWhitelist {
			err = k.WhitelistedPairs.Set(ctx, pair)
			if err != nil {
				panic(err)
			}
		}
	}
}
