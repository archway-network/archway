package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/rewards/types"
)

// SetContractMetadata creates or updates the contract metadata verifying the ownership:
//   - Meta could be created by the contract admin (if set);
//   - Meta could be modified by the contract owner;
func (k Keeper) SetContractMetadata(ctx sdk.Context, senderAddr, contractAddr sdk.AccAddress, metaUpdates types.ContractMetadata) error {
	// Check if the contract exists
	contractInfo := k.contractInfoView.GetContractInfo(ctx, contractAddr)
	if contractInfo == nil {
		return types.ErrContractNotFound
	}

	if metaUpdates.RewardsAddress != "" {
		addr, err := sdk.AccAddressFromBech32(metaUpdates.RewardsAddress)
		if err != nil {
			return err
		}
		if k.isBlockedAddress(addr) {
			return types.ErrInvalidRequest.Wrap("rewards address cannot be a blocked address")
		}
	}

	// Check ownership
	metaOld, err := k.ContractMetadata.Get(ctx, contractAddr)
	if err == nil {
		if metaOld.OwnerAddress != senderAddr.String() {
			return errorsmod.Wrap(types.ErrUnauthorized, "metadata can only be changed by the contract owner")
		}
	} else {
		if contractInfo.Admin != senderAddr.String() {
			return errorsmod.Wrap(types.ErrUnauthorized, "metadata can only be created by the contract admin")
		}
	}

	// Build the updated meta
	metaNew := metaOld
	if err != nil {
		metaNew.ContractAddress = contractAddr.String()
		metaNew.OwnerAddress = senderAddr.String()
	}
	if metaUpdates.HasOwnerAddress() {
		metaNew.OwnerAddress = metaUpdates.OwnerAddress
	}
	if metaUpdates.HasRewardsAddress() {
		metaNew.RewardsAddress = metaUpdates.RewardsAddress
	}
	if metaUpdates.WithdrawToWallet != metaOld.WithdrawToWallet {
		metaNew.WithdrawToWallet = metaUpdates.WithdrawToWallet
	}

	// Set
	err = k.ContractMetadata.Set(ctx, contractAddr, metaNew)
	if err != nil {
		return err
	}

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
	meta, err := k.ContractMetadata.Get(ctx, contractAddr)
	if err != nil {
		return nil
	}

	return &meta
}

func (k Keeper) isBlockedAddress(addr sdk.AccAddress) bool {
	return k.bankKeeper.BlockedAddr(addr)
}
