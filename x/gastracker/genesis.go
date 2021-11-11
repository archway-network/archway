package gastracker

import (
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func InitParams(context sdk.Context, k Keeper) {
	k.SetParams(context, gstTypes.DefaultParams())
}
