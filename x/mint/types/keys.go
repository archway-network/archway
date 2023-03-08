package types

const (
	// ModuleName is the module name.
	ModuleName = "mint"
	// StoreKey is the module KV storage prefix key.
	StoreKey = ModuleName
	// TStoreKey is the module transient storage prefix key.
	TStoreKey = "t_" + ModuleName
	// QuerierRoute is the querier route for the module.
	QuerierRoute = ModuleName
)

// KV Store
var (
	LastBlockInfoPrefix = []byte{0x00}
)

// Transient Store
var (
	MintDistribution = []byte{0x00}
)

// GetValidatorsKey creates the key for the validator with address
func GetMintDistributionKey(recipientName string) []byte {
	return append(MintDistribution, []byte(recipientName)...)
}
