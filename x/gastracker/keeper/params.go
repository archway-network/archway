package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	gstTypes "github.com/archway-network/archway/x/gastracker"
)

func (k Keeper) GetParams(ctx sdk.Context) (params gstTypes.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

func (k Keeper) SetParams(ctx sdk.Context, params gstTypes.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
