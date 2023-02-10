package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/rewards/types"
)

// SetFlatFee checks if a contract has metadata set and stores the given flat fee to be associated with that contract
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

	types.EmitContractFlatFeeSetEvent(
		ctx,
		contractAddr,
		flatFee,
	)
	return nil
}

// GetFlatFee retreives the flat fee stored for a given contract
func (k Keeper) GetFlatFee(ctx sdk.Context, contractAddr sdk.AccAddress) (sdk.Coin, bool) {
	fee, found := k.state.FlatFee(ctx).GetFee(contractAddr)
	if !found {
		return sdk.Coin{}, false
	}

	return fee, true
}

// ImportFlatFees initializes state from the genesis flat fees data.
func (k Keeper) ImportFlatFees(ctx sdk.Context, flatFees []types.FlatFee) {
	for _, flatFee := range flatFees {
		if err := k.SetFlatFee(ctx, flatFee.MustGetContractAddress(), flatFee.GetFlatFee()); err != nil {
			panic(fmt.Sprintf("flat fee: %+v is invalid: %s", flatFee, err))
		}
	}
}

// ExportFlatFees returns the flat fees genesis data for the state.
func (k Keeper) ExportFlatFees(ctx sdk.Context) []types.FlatFee {
	store := prefix.NewStore(k.state.FlatFee(ctx).stateStore, types.FlatFeePrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var fees = make([]types.FlatFee, 0)
	for ; iterator.Valid(); iterator.Next() {
		var coin sdk.Coin
		contractAddr := sdk.AccAddress(iterator.Key())
		k.state.cdc.MustUnmarshal(iterator.Value(), &coin)

		fees = append(fees, types.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         coin,
		})
	}

	return fees
}
