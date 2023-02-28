package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the contract needed to be fulfilled for banking and supply
// dependencies.
type BankKeeper interface {
	// MintCoins creates new coins from thin air and adds it to the module account.
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	GetSupply(ctx sdk.Context, denom string) sdk.Coin
	//SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
}

// StakingKeeper defines the contract needed to be fulfilled for staking and supply
// dependencies.
type StakingKeeper interface {
	// BondedRatio the fraction of the staking tokens which are currently bonded
	BondedRatio(ctx sdk.Context) sdk.Dec
	BondDenom(ctx sdk.Context) string
}
