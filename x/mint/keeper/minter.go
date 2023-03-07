package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetBlockProvisions gets the tokens to be minted in the current block and returns the new inflation amount as well
func (k Keeper) GetBlockProvisions(ctx sdk.Context) (sdk.Coin, sdk.Dec) {
	// todo: put the begin blocker mint amount calculation here
	panic("unimplemented 👻")
}
