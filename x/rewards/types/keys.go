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
)

// ContractMetadata prefixed store state keys.
var (
	// ContractMetadataStatePrefix defines the state global prefix.
	ContractMetadataStatePrefix = []byte{0x00}

	// ContractMetadataPrefix defines the prefix for storing ContractMetadata objects.
	// Key: ContractMetadataStatePrefix | ContractMetadataPrefix | {contractAddress}
	// Value: ContractMetadata
	ContractMetadataPrefix = []byte{0x01}
)
