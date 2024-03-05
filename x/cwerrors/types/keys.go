package types

import (
	"cosmossdk.io/collections"
)

const (
	// ModuleName is the module name.
	ModuleName = "cwerrors"
	// StoreKey is the module KV storage prefix key.
	StoreKey = ModuleName
	// QuerierRoute is the querier route for the module.
	QuerierRoute = ModuleName
)

var (
	ParamsKeyPrefix = collections.NewPrefix(1)
	ErrorsKeyPrefix = collections.NewPrefix(2)
)
