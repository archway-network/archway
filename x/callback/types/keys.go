package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the module name.
	ModuleName = "callback"
	// StoreKey is the module KV storage prefix key.
	StoreKey = ModuleName
	// QuerierRoute is the querier route for the module.
	QuerierRoute = ModuleName
)

var (
	ParamsKey   = []byte{0x01}
	CallbackKey = []byte{0x02}
)

func GetCallbacksByHeightKey(height int64) []byte {
	return append(CallbackKey, byte(height))
}

func GetCallbackByFullyQualifiedKey(height int64, contractAddress sdk.AccAddress, jobID uint64) []byte {
	return append(GetCallbacksByHeightKey(height), append(contractAddress, sdk.Uint64ToBigEndian(jobID)...)...)
}
