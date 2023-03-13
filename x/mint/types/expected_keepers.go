package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the contract needed to be fulfilled for banking and supply
// dependencies.
type BankKeeper interface {
	// MintCoins makes the money printer go brrr and adds it to the module account.
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	// GetSupply retrieves the given token supply from store
	GetSupply(ctx sdk.Context, denom string) sdk.Coin
	// SendCoinsFromModuleToModule sends the given number of coins from one module account to another module account
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
}

// StakingKeeper defines the contract needed to be fulfilled for staking and supply
// dependencies.
type StakingKeeper interface {
	// BondedRatio the fraction of the staking tokens which are currently bonded
	BondedRatio(ctx sdk.Context) sdk.Dec
	// BondDenom - Bondable coin denomination
	BondDenom(ctx sdk.Context) string
}
