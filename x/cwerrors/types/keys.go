package types

import (
	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the module name.
	ModuleName = "cwerrors"
	// StoreKey is the module KV storage prefix key.
	StoreKey = ModuleName
	// QuerierRoute is the querier route for the module.
	QuerierRoute = ModuleName
	// TStoreKey defines the transient store key
	TStoreKey = "t_" + ModuleName
)

// Collections
var (
	// ParamsKeyPrefix is the prefix for the module parameter store.
	ParamsKeyPrefix = collections.NewPrefix(1)
	// ErrorIDKey is the prefix for the count of errors
	ErrorIDKey = collections.NewPrefix(2)
	// ContractErrorsKeyPrefix is the prefix for the collection of all error ids for a given contractzs
	ContractErrorsKeyPrefix = collections.NewPrefix(3)
	// ErrorsKeyPrefix is the prefix for the collection of all errors
	ErrorsKeyPrefix = collections.NewPrefix(4)
	// DeletionBlocksKeyPrefix is the prefix for the collection of all error ids which need to be deleted in given block
	DeletionBlocksKeyPrefix = collections.NewPrefix(5)
	// ContractSubscriptionsKeyPrefix is the prefix for the collection of all contracts with subscriptions
	ContractSubscriptionsKeyPrefix = collections.NewPrefix(6)
	// SubscriptionEndBlockKeyPrefix is the prefix for the collection of all subscriptions which end at given block
	SubscriptionEndBlockKeyPrefix = collections.NewPrefix(7)
)

// Transient Store
var (
	ErrorsForSudoCallbackKey = []byte{0x00}
)

func GetErrorsForSudoCallStoreKey(errorID uint64) []byte {
	return append(ErrorsForSudoCallbackKey, sdk.Uint64ToBigEndian(errorID)...)
}
