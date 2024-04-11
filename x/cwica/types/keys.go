package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name
	ModuleName = "cwica"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	ParamsKeyPrefix = collections.NewPrefix(1)
)
