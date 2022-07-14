package gastracker

const (
	// ModuleName is the name of the contract module
	ModuleName = "gastracker"
	// ContractRewardCollector is the module holding rewards collected by dApps.
	ContractRewardCollector = ModuleName
	// StoreKey is the string store representation
	StoreKey = ModuleName
	// TStoreKey is the string transient store representation
	TStoreKey = "transient_" + ModuleName
	// QuerierRoute is the querier route for the wasm module
	QuerierRoute = ModuleName
	// RouterKey is the msg router key for the wasm module
	RouterKey = ModuleName

	PendingContractInstanceMetadataKeyPrefix = "p_c_inst_md"

	ContractInstanceMetadataKeyPrefix = "c_inst_md"

	RewardEntryKeyPrefix = "reward_entry"
)

var (
	PrefixDappBlockInflationaryRewards = []byte{0x10}
	// PrefixGasTrackingTxIdentifier maps the current transaction being tracked identifier.
	// This value is reset every block. And increased each time keeper.Keeper.TrackNewTx is called.
	PrefixGasTrackingTxIdentifier = []byte{0x11}
	// PrefixGasTrackingTxTracking is the kvstore namespace that contains TransactionTracking objects
	// of the current block.
	PrefixGasTrackingTxTracking = []byte{0x13}

	// KeyTxIdentifier is the constant key used to get the Tx identifier
	// in the current block.
	KeyTxIdentifier = []byte{0x0}
)

func GetPendingContractInstanceMetadataKey(address string) []byte {
	return []byte(PendingContractInstanceMetadataKeyPrefix + "/" + address)
}

func SplitContractAddressFromPendingMetadataKey(key []byte) (contractAddress string) {
	return string(key[len([]byte(PendingContractInstanceMetadataKeyPrefix+"/")):])
}

func GetContractInstanceMetadataKey(address string) []byte {
	return []byte(ContractInstanceMetadataKeyPrefix + "/" + address)
}

func GetRewardEntryKey(rewardAddress string) []byte {
	return []byte(RewardEntryKeyPrefix + "/" + rewardAddress)
}
