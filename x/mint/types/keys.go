package types

const (
	// ModuleName is the module name.
	ModuleName = "mint"
	// StoreKey is the module KV storage prefix key.
	StoreKey = ModuleName
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

func GetMintDistributionRecipientKey(blockHeight int64, recipientName string) []byte {
	return append(append(MintDistribution, byte(blockHeight)), []byte(recipientName)...)
}
