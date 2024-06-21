package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name
	ModuleName = "cwregistry"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	CodeMetadataKeyPrefix = collections.NewPrefix(1)
)
