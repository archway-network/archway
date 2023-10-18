package types

import "cosmossdk.io/collections"

const (
	// ModuleName is the module name.
	ModuleName = "rewards"
	// StoreKey is the module KV storage prefix key.
	StoreKey = ModuleName
	// QuerierRoute is the querier route for the module.
	QuerierRoute = ModuleName
	// RouterKey is the msg router key for the module.
	RouterKey = ModuleName

	// ContractRewardCollector is the module account holding rewards collected by dApps that can be withdrawn.
	ContractRewardCollector = ModuleName

	// TreasuryCollector is the module account name to keep undistributed rewards in.
	TreasuryCollector = "treasury"
)

// Full prefixes
var (
	// ContractMetadataPrefix defines the prefix for storing contract metadata.
	ContractMetadataPrefix = collections.NewPrefix([]byte{0x00, 0x00})
	// BlockRewardsPrefix defines the prefix for storing BlockRewards objects.
	BlockRewardsPrefix = collections.NewPrefix([]byte{0x01, 0x00})
	// TxRewardsPrefix defines the prefix for storing TxRewards objects.
	TxRewardsPrefix = collections.NewPrefix([]byte{0x02, 0x00})
	// TxRewardsHeightIndexPrefix defines the prefix for storing TxRewards's height index.
	TxRewardsHeightIndexPrefix = collections.NewPrefix([]byte{0x02, 0x01})
	// MinConsFeePrefix defines the prefix for storing minimum consensus fee.
	MinConsFeePrefix = collections.NewPrefix([]byte{0x03, 0x00})
	// FlatFeePrefix defines the prefix for storing flat fees.
	FlatFeePrefix = collections.NewPrefix([]byte{0x05, 0x00})
	// ParamsPrefix defines the prefix for storing params.
	ParamsPrefix = collections.NewPrefix([]byte{0x06})
)

// RewardsRecord prefixed store state keys.
var (
	// RewardsRecordStatePrefix defines the state global prefix.
	RewardsRecordStatePrefix = []byte{0x04}

	// RewardsRecordIDKey defines the key for storing last unique RewardsRecord's ID.
	// Key: RewardsRecordStatePrefix | RewardsRecordIDKey
	// Value: uint64
	RewardsRecordIDKey = []byte{0x00}

	// RewardsRecordPrefix defines the prefix for storing RewardsRecord objects.
	// Key: RewardsRecordStatePrefix | RewardsRecordPrefix | {ID}
	// Value: RewardsRecord
	RewardsRecordPrefix = []byte{0x01}

	// RewardsRecordAddressIndexPrefix defines the prefix for storing RewardsRecord's RewardsAddress index.
	// Key: RewardsRecordStatePrefix | RewardsRecordAddressIndexPrefix | {RewardsAddress} | {ID}
	// Value: None
	RewardsRecordAddressIndexPrefix = []byte{0x02}
)
