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

	record, err := k.CreateRewardsRecord(ctx, rewardsAddr, flatfees, calculationHeight, calculationTime)
	if err != nil {
		panic(err)
	}

	// mark the tx as having flat fees rewards records
	err = k.TxFlatFeesIDs.Set(ctx, record.Id)
	if err != nil {
		panic(err)
	}
}

// MaybeReimburseFlatFees checks if the current tx was successful and reimburses the flat fees associated with it.
// It will clear the flat fees ids associated with the current tx.
// If the tx is success nothing further is done, toReimburse will be empty.
// If the tx has failed, then flat fees rewards record associated with the current tx will be cleared
// and reimbursed.
func (k Keeper) MaybeReimburseFlatFees(ctx sdk.Context, txSuccess bool, feePayer sdk.AccAddress) (reimbursed sdk.Coins, err error) {
	var rewardsRecordsID []uint64
	err = k.TxFlatFeesIDs.Walk(ctx, nil, func(id uint64) (stop bool, err error) {
		rewardsRecordsID = append(rewardsRecordsID, id)
		return false, nil
	})
	if err != nil {
		return sdk.Coins{}, err
	}
	// no flat fees rewards records, means no flat fees to reimburse.
	if len(rewardsRecordsID) == 0 {
		return sdk.NewCoins(), nil
	}
	// we clear the records associated with the current tx
	err = k.TxFlatFeesIDs.Clear(ctx, nil)
	if err != nil {
		return sdk.Coins{}, err
	}
	// if tx went well then there's nothing to reimburse, we simply clear the records
	// associated with the current tx.
	if txSuccess {
		return sdk.NewCoins(), nil
	}
	// if tx failed then we reimburse the flat fees associated with the current tx
	reimbursed = sdk.NewCoins()
	for _, rewardRecordID := range rewardsRecordsID {
		rewardRecord, err := k.RewardsRecords.Get(ctx, rewardRecordID)
		if err != nil {
			return sdk.Coins{}, err
		}
		reimbursed = reimbursed.Add(rewardRecord.Rewards...)
		// delete reward record
		err = k.RewardsRecords.Remove(ctx, rewardRecordID)
		if err != nil {
			return sdk.Coins{}, err
		}
	}
	// reimburse
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, feePayer, reimbursed)
	if err != nil {
		return sdk.Coins{}, err
	}
	return reimbursed, nil
}
