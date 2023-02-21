package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/archway-network/archway/x/rewards/types"
)

// SetFlatFee checks if a contract has metadata set and stores the given flat fee to be associated with that contract
func (k Keeper) SetFlatFee(ctx sdk.Context, senderAddr sdk.AccAddress, feeUpdate types.FlatFee) error {
	// Check if the contract metadata exists
	contractInfo := k.GetContractMetadata(ctx, feeUpdate.MustGetContractAddress())
	if contractInfo == nil {
		return types.ErrMetadataNotFound
	}
	if contractInfo.OwnerAddress != senderAddr.String() {
		return sdkErrors.Wrap(types.ErrUnauthorized, "flat_fee can only be set or changed by the contract owner")
	}

	if feeUpdate.FlatFee.Amount.IsZero() {
		k.state.FlatFee(ctx).RemoveFee(feeUpdate.MustGetContractAddress())
	} else {
		k.state.FlatFee(ctx).SetFee(feeUpdate.MustGetContractAddress(), feeUpdate.FlatFee)
	}

	types.EmitContractFlatFeeSetEvent(
		ctx,
		feeUpdate.MustGetContractAddress(),
		feeUpdate.FlatFee,
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
