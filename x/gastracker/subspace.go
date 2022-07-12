package gastracker

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsTypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type Subspace interface {
	Get(ctx sdk.Context, key []byte, ptr interface{})
	SetParamSet(ctx sdk.Context, params paramsTypes.ParamSet)
	GetParamSet(ctx sdk.Context, ps paramsTypes.ParamSet)
}
