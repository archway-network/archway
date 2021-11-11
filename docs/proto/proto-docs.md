<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [gastracker/types.proto](#gastracker/types.proto)
    - [BlockGasTracking](#gastracker.BlockGasTracking)
    - [ContractGasTracking](#gastracker.ContractGasTracking)
    - [ContractInstanceMetadata](#gastracker.ContractInstanceMetadata)
    - [ContractInstantiationRequestWrapper](#gastracker.ContractInstantiationRequestWrapper)
    - [ContractOperationInfo](#gastracker.ContractOperationInfo)
    - [GasTrackingQueryRequestWrapper](#gastracker.GasTrackingQueryRequestWrapper)
    - [GasTrackingQueryResultWrapper](#gastracker.GasTrackingQueryResultWrapper)
    - [LeftOverRewardEntry](#gastracker.LeftOverRewardEntry)
    - [RewardDistributionEvent](#gastracker.RewardDistributionEvent)
    - [TransactionTracking](#gastracker.TransactionTracking)
  
    - [ContractOperation](#gastracker.ContractOperation)
  
- [Scalar Value Types](#scalar-value-types)



<a name="gastracker/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## gastracker/types.proto



<a name="gastracker.BlockGasTracking"></a>

### BlockGasTracking
Tracking gas consumption for all tx in a particular block


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_tracking_infos` | [TransactionTracking](#gastracker.TransactionTracking) | repeated |  |






<a name="gastracker.ContractGasTracking"></a>

### ContractGasTracking
Tracking contract gas usage


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `gas_consumed` | [uint64](#uint64) |  |  |
| `is_eligible_for_reward` | [bool](#bool) |  |  |
| `operation` | [ContractOperation](#gastracker.ContractOperation) |  |  |






<a name="gastracker.ContractInstanceMetadata"></a>

### ContractInstanceMetadata
Contract instance metadata


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `reward_address` | [string](#string) |  |  |
| `gas_rebate_to_user` | [bool](#bool) |  |  |
| `collect_premium` | [bool](#bool) |  | Flag to indicate whether to charge premium or not |
| `premium_percentage_charged` | [uint64](#uint64) |  | Percentage of gas consumed to be charged. |






<a name="gastracker.ContractInstantiationRequestWrapper"></a>

### ContractInstantiationRequestWrapper
Custom wrapper around contract instantiation request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `reward_address` | [string](#string) |  |  |
| `gas_rebate_to_user` | [bool](#bool) |  |  |
| `collect_premium` | [bool](#bool) |  |  |
| `premium_percentage_charged` | [uint64](#uint64) |  |  |
| `instantiation_request` | [string](#string) |  | Base64 encoding of instantiation data |






<a name="gastracker.ContractOperationInfo"></a>

### ContractOperationInfo
Custom Message returned by our wrapper vm


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gas_consumed` | [uint64](#uint64) |  |  |
| `operation` | [ContractOperation](#gastracker.ContractOperation) |  |  |
| `reward_address` | [string](#string) |  | Only set in case of instantiate operation |
| `gas_rebate_to_end_user` | [bool](#bool) |  | Only set in case of instantiate operation |
| `collect_premium` | [bool](#bool) |  | Only set in case of instantiate operation |
| `premium_percentage_charged` | [uint64](#uint64) |  | Only set in case of instantiate operation |






<a name="gastracker.GasTrackingQueryRequestWrapper"></a>

### GasTrackingQueryRequestWrapper
Custom wrapper around Query request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `magic_string` | [string](#string) |  |  |
| `query_request` | [bytes](#bytes) |  |  |






<a name="gastracker.GasTrackingQueryResultWrapper"></a>

### GasTrackingQueryResultWrapper
Custom wrapper around Query result that also gives gas consumption


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gas_consumed` | [uint64](#uint64) |  |  |
| `query_response` | [bytes](#bytes) |  |  |






<a name="gastracker.LeftOverRewardEntry"></a>

### LeftOverRewardEntry
Reward entry per beneficiary address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_rewards` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated |  |






<a name="gastracker.RewardDistributionEvent"></a>

### RewardDistributionEvent
Event emitted when Reward is allocated


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `reward_address` | [string](#string) |  |  |
| `contract_rewards` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
| `leftover_rewards` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated |  |






<a name="gastracker.TransactionTracking"></a>

### TransactionTracking
Tracking contract gas usage and total gas consumption per transaction


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `max_gas_allowed` | [uint64](#uint64) |  |  |
| `max_contract_rewards` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated |  |
| `contract_tracking_infos` | [ContractGasTracking](#gastracker.ContractGasTracking) | repeated |  |





 <!-- end messages -->


<a name="gastracker.ContractOperation"></a>

### ContractOperation
Denotes which operation consumed this gas

| Name | Number | Description |
| ---- | ------ | ----------- |
| CONTRACT_OPERATION_UNSPECIFIED | 0 | Invalid or unknown operation |
| CONTRACT_OPERATION_INSTANTIATION | 1 | Initialization of the contract |
| CONTRACT_OPERATION_EXECUTION | 2 | Execution of the contract |
| CONTRACT_OPERATION_QUERY | 3 | Querying the contract |
| CONTRACT_OPERATION_MIGRATE | 4 | Migrating the contract |
| CONTRACT_OPERATION_IBC | 5 | IBC operation |
| CONTRACT_OPERATION_SUDO | 6 | Sudo operation |
| CONTRACT_OPERATION_REPLY | 7 | Reply operation |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

