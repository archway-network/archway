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
	MintDistribution    = []byte{0x01}
)

// GetMintDistributionRecipientKey gets the store prefix to fetch the inflation distribution for the recipient
// returns MintDistribution + currentBlockHeight + recipientName
func GetMintDistributionRecipientKey(blockHeight int64, recipientName string) []byte {
	return append(append(MintDistribution, byte(blockHeight)), []byte(recipientName)...)
}
