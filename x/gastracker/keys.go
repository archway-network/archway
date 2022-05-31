package gastracker

const (
	// ModuleName is the name of the contract module
	ModuleName = "gastracker"

	GasRewardCollector = ModuleName

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// TStoreKey is the string transient store representation
	TStoreKey = "transient_" + ModuleName

	// QuerierRoute is the querier route for the wasm module
	QuerierRoute = ModuleName

	// RouterKey is the msg router key for the wasm module
	RouterKey = ModuleName

	CurrentBlockTrackingKey = "current_blk"

	PendingContractInstanceMetadataKeyPrefix = "p_c_inst_md"

	ContractInstanceMetadataKeyPrefix = "c_inst_md"

	ContractInstanceSystemMetadataKeyPrefix = "c_inst_smd"

	RewardEntryKeyPrefix = "reward_entry"

	GlobalTxCounterKey = "gtc"

	InflationRewardAccumulator = ModuleName + "_accumulator"
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

func GetContractInstanceSystemMetadataKey(address string) []byte {
	return []byte(ContractInstanceSystemMetadataKeyPrefix + "/" + address)
}
