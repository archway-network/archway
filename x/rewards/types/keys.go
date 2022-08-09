package types

const (
	// ModuleName is the module name.
	ModuleName = "rewards"
	// StoreKey is the module KV storage prefix key.
	StoreKey = ModuleName
	// QuerierRoute is the querier route for the module.
	QuerierRoute = ModuleName
	// RouterKey is the msg router key for the module.
	RouterKey = ModuleName

	// ContractRewardCollector is the module holding rewards collected by dApps.
	ContractRewardCollector = ModuleName
)

// ContractMetadata prefixed store state keys.
var (
	// ContractMetadataStatePrefix defines the state global prefix.
	ContractMetadataStatePrefix = []byte{0x00}

	// ContractMetadataPrefix defines the prefix for storing ContractMetadata objects.
	// Key: ContractMetadataStatePrefix | ContractMetadataPrefix | {ContractAddress}
	// Value: ContractMetadata
	ContractMetadataPrefix = []byte{0x00}
)

// BlockRewards prefixed store state keys.
var (
	// BlockRewardsStatePrefix defines the state global prefix.
	BlockRewardsStatePrefix = []byte{0x01}

	// BlockRewardsPrefix defines the prefix for storing BlockRewards objects.
	// Key: BlockRewardsStatePrefix | BlockRewardsPrefix | {Height}
	// Value: BlockRewards
	BlockRewardsPrefix = []byte{0x00}
)

// TxRewards prefixed store state keys.
var (
	// TxRewardsStatePrefix defines the state global prefix.
	TxRewardsStatePrefix = []byte{0x02}

	// TxRewardsPrefix defines the prefix for storing TxRewards objects.
	// Key: TxRewardsStatePrefix | TxRewardsPrefix | {TxID}
	// Value: TxRewards
	TxRewardsPrefix = []byte{0x00}

	// TxRewardsBlockIndexPrefix defines the prefix for storing TxRewards's block index.
	// Key: TxRewardsStatePrefix | TxRewardsBlockIndexPrefix | {Height} | {TxID}
	// Value: None
	TxRewardsBlockIndexPrefix = []byte{0x01}
)

// Minimum consensus fee store state keys.
var (
	// MinConsFeeStatePrefix defines the state global prefix.
	MinConsFeeStatePrefix = []byte{0x03}

	// MinConsFeeKey defines the key for storing MinConsFee coin.
	// Key: MinConsFeeStatePrefix | MinConsFeeKey
	// Value: sdk.Coin
	MinConsFeeKey = []byte{0x00}
)
