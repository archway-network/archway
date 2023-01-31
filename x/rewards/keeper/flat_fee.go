package keeper

import (
	"github.com/archway-network/archway/x/rewards/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetFlatFee
func (k Keeper) SetFlatFee(ctx sdk.Context, contractAddr sdk.AccAddress, flatFee sdk.Coin) error {
	// Check if the contract metadata exists
	contractInfo := k.GetContractMetadata(ctx, contractAddr)
	if contractInfo == nil {
		return types.ErrMetadataNotFound
	}

	if flatFee.Amount.IsZero() {
		k.state.FlatFee(ctx).RemoveFee(contractAddr)
	} else {
		k.state.FlatFee(ctx).SetFee(contractAddr, flatFee)
	}
	return nil
}

// GetFlatFee
func (k Keeper) GetFlatFee(ctx sdk.Context, contractAddr sdk.AccAddress) (sdk.Coin, bool) {
	fee, found := k.state.FlatFee(ctx).GetFee(contractAddr)
	if !found {
		return sdk.Coin{}, false
	}

	return fee, true
}
