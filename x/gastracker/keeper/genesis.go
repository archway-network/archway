package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	gstTypes "github.com/archway-network/archway/x/gastracker"
)

func InitParams(context sdk.Context, k GasTrackingKeeper) {
	k.SetParams(context, gstTypes.DefaultParams())
}
