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
	// RewardsRecordsIDPrefix defines the prefix for storing RewardsRecord last ID.
	RewardsRecordsIDPrefix = collections.NewPrefix([]byte{0x04, 0x00})
	// RewwardsRecordStatePrefix defines the prefix for storing RewardsRecord state.
	RewardsRecordStatePrefix = collections.NewPrefix([]byte{0x04, 0x01})
	// RewardsRecordAddressIndexPrefix defines the prefix for storing RewardsRecord's rewards address index.
	RewardsRecordAddressIndexPrefix = collections.NewPrefix([]byte{0x04, 0x02})
	// FlatFeePrefix defines the prefix for storing flat fees.
	FlatFeePrefix = collections.NewPrefix([]byte{0x05, 0x00})
	// ParamsPrefix defines the prefix for storing params.
	ParamsPrefix = collections.NewPrefix([]byte{0x06})
	// TxFlatFeesIDsPrefix defines the prefix for storing TxFlatFees last IDs.
	TxFlatFeesIDsPrefix = collections.NewPrefix([]byte{0x07})
)
