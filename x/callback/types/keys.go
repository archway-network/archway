package types

const (
	// ModuleName is the module name.
	ModuleName = "callback"
	// StoreKey is the module KV storage prefix key.
	StoreKey = ModuleName
	// QuerierRoute is the querier route for the module.
	QuerierRoute = ModuleName
)

var (
	ParamsKey = []byte{0x01}
)
