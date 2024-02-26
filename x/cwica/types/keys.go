package types

const (
	// ModuleName defines the module name
	ModuleName = "cwica"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_cwica"
)

const (
	// parameters key
	prefixParamsKey = iota + 1
)

var (
	ParamsKey = []byte{prefixParamsKey}
)
