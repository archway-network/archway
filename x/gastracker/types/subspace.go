package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Subspace interface {
	Get(ctx sdk.Context, key []byte, ptr interface{}) bool
	SetParamSet(ctx sdk.Context, params Params)
}
