package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type WasmKeeperExpected interface {
	HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool
	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
}
