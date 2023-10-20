package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
		return errorsmod.Wrap(types.ErrUnauthorized, "flat_fee can only be set or changed by the contract owner")
	}

	if feeUpdate.FlatFee.Amount.IsZero() {
		err := k.FlatFees.Remove(ctx, feeUpdate.MustGetContractAddress())
		if err != nil {
			return err
		}
	} else {
		if contractInfo.RewardsAddress == "" {
			return errorsmod.Wrap(types.ErrMetadataNotFound, "flat_fee can only be set when rewards address has been configured")
		}
		err := k.FlatFees.Set(ctx, feeUpdate.MustGetContractAddress(), feeUpdate.FlatFee)
		if err != nil {
			return err
		}
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
	fee, err := k.FlatFees.Get(ctx, contractAddr)
	if err != nil {
		return sdk.Coin{}, false
	}

	return fee, true
}

// CreateFlatFeeRewardsRecords creates a rewards record for the flatfees of the given contract
func (k Keeper) CreateFlatFeeRewardsRecords(ctx sdk.Context, contractAddress sdk.AccAddress, flatfees sdk.Coins) {
	calculationHeight, calculationTime := ctx.BlockHeight(), ctx.BlockTime()

	metadata := k.GetContractMetadata(ctx, contractAddress)
	rewardsAddr := sdk.MustAccAddressFromBech32(metadata.RewardsAddress)

	_, err := k.CreateRewardsRecord(ctx, rewardsAddr, flatfees, calculationHeight, calculationTime)
	if err != nil {
		panic(err)
	}
}
