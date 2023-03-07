package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the contract needed to be fulfilled for banking and supply
// dependencies.
type BankKeeper interface {
	// MintCoins makes the money printer go brrr and adds it to the module account.
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error
}
