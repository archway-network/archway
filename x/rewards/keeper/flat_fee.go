package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetFlatFee
func (k Keeper) SetFlatFee(ctx sdk.Context, contractAddr sdk.AccAddress, flatFee sdk.Coin) {
	if flatFee.Amount.IsZero() {
		k.state.FlatFee(ctx).RemoveFee(contractAddr)
	} else {
		k.state.FlatFee(ctx).SetFee(contractAddr, flatFee)
	}
}

// GetFlatFee
func (k Keeper) GetFlatFee(ctx sdk.Context, contractAddr sdk.AccAddress) (sdk.Coin, bool) {
	fee, found := k.state.FlatFee(ctx).GetFee(contractAddr)
	if !found {
		return sdk.Coin{}, false
	}

	return fee, true
}
