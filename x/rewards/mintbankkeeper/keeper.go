package mintbankkeeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/archway-network/archway/pkg"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

var _ mintTypes.BankKeeper = Keeper{}

// RewardsKeeperExpected defines the expected interface for the x/rewards keeper.
type RewardsKeeperExpected interface {
	InflationRewardsRatio(ctx sdk.Context) sdk.Dec
	TrackInflationRewards(ctx sdk.Context, rewards sdk.Coin)
	UpdateMinConsensusFee(ctx sdk.Context, inflationRewards sdk.Coin)
}

// Keeper is the x/bank keeper decorator that is used by the x/mint module.
// Decorator is used to split inflation tokens between the rewards collector and the fee collector accounts.
type Keeper struct {
	bankKeeper    mintTypes.BankKeeper
	rewardsKeeper RewardsKeeperExpected
}

// NewKeeper creates a new Keeper instance.
func NewKeeper(bk mintTypes.BankKeeper, rk RewardsKeeperExpected) Keeper {
	return Keeper{
		bankKeeper:    bk,
		rewardsKeeper: rk,
	}
}

// SendCoinsFromModuleToModule implements the mintTypes.BankKeeper interface.
func (k Keeper) SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error {
	// Perform the split only if the recipient is fee collector (which for instance is always the case) and
	// inflation rewards are enabled.
	ratio := k.rewardsKeeper.InflationRewardsRatio(ctx)
	if recipientModule != authTypes.FeeCollectorName || ratio.IsZero() {
		return k.bankKeeper.SendCoinsFromModuleToModule(ctx, senderModule, recipientModule, amt)
	}

	dappRewards, stakingRewards := pkg.SplitCoins(amt, ratio)

	// Send to the x/auth fee collector account
	if !stakingRewards.Empty() {
		if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, senderModule, recipientModule, stakingRewards); err != nil {
			return err
		}
	}

	// Send to the x/rewards account
	if !dappRewards.Empty() {
		if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, senderModule, rewardsTypes.ContractRewardCollector, dappRewards); err != nil {
			return err
		}
	}

	// Check that only one coin has been minted
	if len(dappRewards) != 1 {
		panic(fmt.Errorf("unexpected dApp rewards: %s", dappRewards))
	}

	// Track inflation rewards
	k.rewardsKeeper.TrackInflationRewards(ctx, dappRewards[0])
	// Update the minimum consensus fee
	k.rewardsKeeper.UpdateMinConsensusFee(ctx, dappRewards[0])

	return nil
}

// SendCoinsFromModuleToAccount implements the mintTypes.BankKeeper interface.
func (k Keeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, amt)
}

// MintCoins implements the mintTypes.BankKeeper interface.
func (k Keeper) MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error {
	return k.bankKeeper.MintCoins(ctx, name, amt)
}
