package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/archway-network/archway/x/rewards/types"
)

// SetContractMetadata creates or updates the contract metadata verifying the ownership:
//   - Meta could be created by the contract admin (if set);
//   - Meta could be modified by the contract owner;
func (k Keeper) SetContractMetadata(ctx sdk.Context, senderAddr, contractAddr sdk.AccAddress, metaUpdates types.ContractMetadata) error {
	state := k.state.ContractMetadataState(ctx)

	// Check if the contract exists
	contractInfo := k.contractInfoView.GetContractInfo(ctx, contractAddr)
	if contractInfo == nil {
		return types.ErrContractNotFound
	}

	// Check ownership
	metaOld, metaExists := state.GetContractMetadata(contractAddr)
	if metaExists {
		if metaOld.OwnerAddress != senderAddr.String() {
			return sdkErrors.Wrap(types.ErrUnauthorized, "metadata can only be changed by the contract owner")
		}
	} else {
		if contractInfo.Admin != senderAddr.String() {
			return sdkErrors.Wrap(types.ErrUnauthorized, "metadata can only be created by the contract admin")
		}
	}

	// Build the updated meta
	metaNew := metaOld
	if !metaExists {
		metaNew.ContractAddress = contractAddr.String()
		metaNew.OwnerAddress = senderAddr.String()
	}
	if metaUpdates.HasOwnerAddress() {
		metaNew.OwnerAddress = metaUpdates.OwnerAddress
	}
	if metaUpdates.HasRewardsAddress() {
		metaNew.RewardsAddress = metaUpdates.RewardsAddress
	}

	// Set
	state.SetContractMetadata(contractAddr, metaNew)

	// Emit event
	types.EmitContractMetadataSetEvent(
		ctx,
		contractAddr,
		metaNew,
	)

	return nil
}

// GetContractMetadata returns the contract metadata for the given contract address (if found).
func (k Keeper) GetContractMetadata(ctx sdk.Context, contractAddr sdk.AccAddress) *types.ContractMetadata {
	meta, found := k.state.ContractMetadataState(ctx).GetContractMetadata(contractAddr)
	if !found {
		return nil
	}

	return &meta
}
