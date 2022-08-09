package mintbankkeeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/archway-network/archway/x/gastracker"
	"github.com/archway-network/archway/x/gastracker/common"
)

var (
	_ minttypes.BankKeeper = (*Keeper)(nil)
)

type GasTrackingKeeper interface {
	GetParams(ctx sdk.Context) gastracker.Params
	UpdateDappInflationaryRewards(ctx sdk.Context, rewards sdk.Coin)
}

func NewKeeper(bk minttypes.BankKeeper, gtk GasTrackingKeeper) Keeper {
	return Keeper{
		bk:  bk,
		gtk: gtk,
	}
}

// Keeper mocks the behaviour of the bank keeper required
// by the mint module and splits inflationary rewards
// between the gas tracking module and auth's fee collector
type Keeper struct {
	bk  minttypes.BankKeeper
	gtk GasTrackingKeeper
}

// SendCoinsFromModuleToModule overrides the behaviour of mint's BankKeeper and redirects part of inflationary
// rewards towards the gastracker module.
func (k Keeper) SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error {
	// we only care about this if the recipient is fee collector
	// (which for instance is always the case)
	if recipientModule != authtypes.FeeCollectorName {
		return k.bk.SendCoinsFromModuleToModule(ctx, senderModule, recipientModule, amt)
	}

	ratio := k.gtk.GetParams(ctx).DappInflationRewardsRatio
	stakingRewards, dappRewards := common.SplitCoins(ratio, amt)

	// send to auth's fee collector
	err := k.bk.SendCoinsFromModuleToModule(ctx, senderModule, recipientModule, stakingRewards)
	if err != nil {
		return err
	}
	// send to gastracker
	err = k.bk.SendCoinsFromModuleToModule(ctx, senderModule, gastracker.ContractRewardCollector, dappRewards)
	if err != nil {
		return err
	}

	if len(dappRewards) != 1 {
		panic(fmt.Errorf("unexpected dapp rewards: %s", dappRewards))
	}
	k.gtk.UpdateDappInflationaryRewards(ctx, dappRewards[0]) // note the minted coin is only and always one

	return nil
}

func (k Keeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return k.bk.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, amt)
}

func (k Keeper) MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error {
	return k.bk.MintCoins(ctx, name, amt)
}
