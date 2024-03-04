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
)

const (
	// params key
	prefixParamsKey = iota + 1
)

var (
	// params store key
	ParamsKey = []byte{prefixParamsKey}
)
