package types

const (
	// ModuleName is the name of the contract module
	ModuleName = "gastracker"

	ContractRewardCollector = ModuleName

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// TStoreKey is the string transient store representation
	TStoreKey = "transient_" + ModuleName

	// QuerierRoute is the querier route for the wasm module
	QuerierRoute = ModuleName

	// RouterKey is the msg router key for the wasm module
	RouterKey = ModuleName

	CurrentBlockTrackingKey = "current_blk"

	ContractInstanceMetadataKeyPrefix = "c_inst_md"

	RewardEntryKeyPrefix = "reward_entry"

	MagicString = "TjWnZr4u7x!A%D*G-KaPdSgUkXp2s5v8y/B?E(H+MbQeThWmYq3t6w9z$C&F)J@N"

	GasTrackingQueryRequestMagicString = MagicString

	GasRebateToUserDescriptor = "SmartContractGasRebateToUser"

	PremiumGasDescriptor = "SmartContractPremiumGas"
)

func GetContractInstanceMetadataKey(address string) []byte {
	return []byte(ContractInstanceMetadataKeyPrefix + "/" + address)
}

func GetRewardEntryKey(rewardAddress string) []byte {
	return []byte(RewardEntryKeyPrefix + "/" + rewardAddress)
}