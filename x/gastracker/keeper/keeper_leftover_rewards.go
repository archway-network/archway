package keeper

import (
	"github.com/archway-network/archway/x/gastracker"
	"github.com/cosmos/cosmos-sdk/types"
)

// TODO(fdymylja): this seems highly unnecessary as coins are defined in micro-units 10^6
// this means that left-overs can be safely ignored as even over time they accrue to meaningless values.

func (k Keeper) CreateOrMergeLeftOverRewardEntry(ctx types.Context, rewardAddress types.AccAddress, contractRewards types.DecCoins, leftOverThreshold uint64) (types.Coins, error) {
	contractRewards = contractRewards.Sort()

	gstKvStore := ctx.KVStore(k.key)

	var rewardEntry gastracker.LeftOverRewardEntry
	var updatedRewards types.DecCoins
	var rewardsToBeDistributed types.Coins

	bz := gstKvStore.Get(gastracker.GetRewardEntryKey(rewardAddress.String()))
	if bz != nil {
		err := k.cdc.Unmarshal(bz, &rewardEntry)
		if err != nil {
			return rewardsToBeDistributed, err
		}
		previousRewards := make(types.DecCoins, len(rewardEntry.ContractRewards))
		for i := range previousRewards {
			previousRewards[i] = *rewardEntry.ContractRewards[i]
		}
		updatedRewards = previousRewards.Add(contractRewards...)
	} else {
		updatedRewards = contractRewards
	}

	rewardsToBeDistributed = make(types.Coins, len(updatedRewards))
	distributionRewardIndex := 0

	leftOverContractRewards := make(types.DecCoins, len(updatedRewards))
	leftOverRewardIndex := 0

	leftOverDec := types.NewDecFromBigInt(gastracker.ConvertUint64ToBigInt(leftOverThreshold))

	for i := range updatedRewards {
		if updatedRewards[i].Amount.GTE(leftOverDec) {
			distributionAmount := updatedRewards[i].Amount.TruncateInt()
			leftOverAmount := updatedRewards[i].Amount.Sub(distributionAmount.ToDec())
			if !leftOverAmount.IsZero() {
				leftOverContractRewards[leftOverRewardIndex] = types.NewDecCoinFromDec(updatedRewards[i].Denom, leftOverAmount)
				leftOverRewardIndex += 1
			}
			rewardsToBeDistributed[distributionRewardIndex] = types.NewCoin(updatedRewards[i].Denom, distributionAmount)
			distributionRewardIndex += 1
		} else {
			leftOverContractRewards[leftOverRewardIndex] = updatedRewards[i]
			leftOverRewardIndex += 1
		}
	}

	rewardsToBeDistributed = rewardsToBeDistributed[:distributionRewardIndex]
	leftOverContractRewards = leftOverContractRewards[:leftOverRewardIndex]

	rewardEntry.ContractRewards = make([]*types.DecCoin, len(leftOverContractRewards))
	for i := range leftOverContractRewards {
		rewardEntry.ContractRewards[i] = &leftOverContractRewards[i]
	}

	bz, err := k.cdc.Marshal(&rewardEntry)
	if err != nil {
		return rewardsToBeDistributed, err
	}

	gstKvStore.Set(gastracker.GetRewardEntryKey(rewardAddress.String()), bz)
	return rewardsToBeDistributed, nil
}

// Since we can only transfer integer numbers
// and rewards can be floating point numbers,
// we accumulate all the rewards and once it reaches to
// an integer number, we pay the integer part and
// keep the 0.x amount as left over to be paid later
func (k Keeper) GetLeftOverRewardEntry(ctx types.Context, rewardAddress types.AccAddress) (gastracker.LeftOverRewardEntry, error) {
	gstKvStore := ctx.KVStore(k.key)

	var rewardEntry gastracker.LeftOverRewardEntry

	bz := gstKvStore.Get(gastracker.GetRewardEntryKey(rewardAddress.String()))
	if bz == nil {
		return rewardEntry, gastracker.ErrRewardEntryNotFound
	}

	err := k.cdc.Unmarshal(bz, &rewardEntry)
	if err != nil {
		return rewardEntry, err
	}

	return rewardEntry, nil
}
