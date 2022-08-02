package types

const (
	// ModuleName is the module name.
	ModuleName = "tracking"
	// StoreKey is the module KV storage prefix key.
	StoreKey = ModuleName
	// QuerierRoute is the querier route for the module.
	QuerierRoute = ModuleName
)

// TxInfo prefixed store state keys.
var (
	// TxInfoStatePrefix defines the state global prefix.
	TxInfoStatePrefix = []byte{0x00}

	// TxInfoIDKey defines the key for storing last unique TxInfo's ID.
	// Key: TxInfoStatePrefix | TxInfoIDKey
	// Value: uint64
	TxInfoIDKey = []byte{0x00}

	// TxInfoPrefix defines the prefix for storing TxInfo objects.
	// Key: TxInfoStatePrefix | TxInfoPrefix | {ID}
	// Value: TxInfo
	TxInfoPrefix = []byte{0x01}

	// TxInfoBlockIndexPrefix defines the prefix for storing TxInfo's block index.
	// Key: TxInfoStatePrefix | TxInfoBlockIndexPrefix | {Height} | {ID}
	// Value: None
	TxInfoBlockIndexPrefix = []byte{0x02}
)

// ContractOperationInfo prefixed store state keys.
var (
	// ContractOpInfoStatePrefix defines the state global prefix.
	ContractOpInfoStatePrefix = []byte{0x01}

	// ContractOpInfoIDKey defines the key for storing last unique ContractOperationInfo's ID.
	// Key: ContractOpInfoStatePrefix | ContractOpInfoIDKey
	// Value: uint64
	ContractOpInfoIDKey = []byte{0x00}

	// ContractOpInfoPrefix defines the prefix for storing ContractOperationInfo objects.
	// Key: ContractOpInfoStatePrefix | ContractOpInfoPrefix | {ID}
	// Value: ContractOperationInfo
	ContractOpInfoPrefix = []byte{0x01}

	// ContractOpInfoTxIndexPrefix defines the prefix for storing ContractOperationInfo's TxInfo index.
	// Key: ContractOpInfoStatePrefix | ContractOpInfoTxIndexPrefix | {TxInfoID} | {ID}
	// Value: None
	ContractOpInfoTxIndexPrefix = []byte{0x02}
)
