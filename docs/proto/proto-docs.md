<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [archway/gastracker/v1/types.proto](#archway/gastracker/v1/types.proto)
    - [BlockGasTracking](#archway.gastracker.v1.BlockGasTracking)
    - [ContractGasTracking](#archway.gastracker.v1.ContractGasTracking)
    - [ContractInstanceMetadata](#archway.gastracker.v1.ContractInstanceMetadata)
    - [ContractInstanceSystemMetadata](#archway.gastracker.v1.ContractInstanceSystemMetadata)
    - [ContractRewardCalculationEvent](#archway.gastracker.v1.ContractRewardCalculationEvent)
    - [ContractValidFeeGranteeMsg](#archway.gastracker.v1.ContractValidFeeGranteeMsg)
    - [GenesisState](#archway.gastracker.v1.GenesisState)
    - [LeftOverRewardEntry](#archway.gastracker.v1.LeftOverRewardEntry)
    - [RewardDistributionEvent](#archway.gastracker.v1.RewardDistributionEvent)
    - [TransactionTracking](#archway.gastracker.v1.TransactionTracking)
    - [ValidateFeeGrant](#archway.gastracker.v1.ValidateFeeGrant)
    - [WasmMsg](#archway.gastracker.v1.WasmMsg)
  
    - [ContractOperation](#archway.gastracker.v1.ContractOperation)
    - [WasmMsgType](#archway.gastracker.v1.WasmMsgType)
  
- [archway/gastracker/v1/query.proto](#archway/gastracker/v1/query.proto)
    - [QueryBlockGasTrackingRequest](#archway.gastracker.v1.QueryBlockGasTrackingRequest)
    - [QueryBlockGasTrackingResponse](#archway.gastracker.v1.QueryBlockGasTrackingResponse)
    - [QueryContractMetadataRequest](#archway.gastracker.v1.QueryContractMetadataRequest)
    - [QueryContractMetadataResponse](#archway.gastracker.v1.QueryContractMetadataResponse)
  
    - [Query](#archway.gastracker.v1.Query)
  
- [archway/gastracker/v1/tx.proto](#archway/gastracker/v1/tx.proto)
    - [MsgSetContractMetadata](#archway.gastracker.v1.MsgSetContractMetadata)
    - [MsgSetContractMetadataResponse](#archway.gastracker.v1.MsgSetContractMetadataResponse)
  
    - [Msg](#archway.gastracker.v1.Msg)
  
- [Scalar Value Types](#scalar-value-types)



<a name="archway/gastracker/v1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## archway/gastracker/v1/types.proto



<a name="archway.gastracker.v1.BlockGasTracking"></a>

### BlockGasTracking
Tracking gas consumption for all tx in a particular block


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_tracking_infos` | [TransactionTracking](#archway.gastracker.v1.TransactionTracking) | repeated |  |






<a name="archway.gastracker.v1.ContractGasTracking"></a>

### ContractGasTracking
Tracking contract gas usage


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `original_vm_gas` | [uint64](#uint64) |  |  |
| `original_sdk_gas` | [uint64](#uint64) |  |  |
| `operation` | [ContractOperation](#archway.gastracker.v1.ContractOperation) |  |  |






<a name="archway.gastracker.v1.ContractInstanceMetadata"></a>

### ContractInstanceMetadata
Contract instance metadata


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `developer_address` | [string](#string) |  | Developer address of the contract |
| `reward_address` | [string](#string) |  |  |
| `gas_rebate_to_user` | [bool](#bool) |  |  |
| `collect_premium` | [bool](#bool) |  | Flag to indicate whether to charge premium or not |
| `premium_percentage_charged` | [uint64](#uint64) |  | Percentage of gas consumed to be charged. |






<a name="archway.gastracker.v1.ContractInstanceSystemMetadata"></a>

### ContractInstanceSystemMetadata
Contract instance system level metadata, not updatable externally.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `inflation_balance` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated | Inflation reward balance of this contract instance. |
| `gas_counter` | [bytes](#bytes) |  |  |






<a name="archway.gastracker.v1.ContractRewardCalculationEvent"></a>

### ContractRewardCalculationEvent
Event emitted when contract reward is calculated


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  |  |
| `gas_consumed` | [uint64](#uint64) |  |  |
| `inflation_rewards` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated |  |
| `contract_rewards` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated |  |
| `metadata` | [ContractInstanceMetadata](#archway.gastracker.v1.ContractInstanceMetadata) |  |  |






<a name="archway.gastracker.v1.ContractValidFeeGranteeMsg"></a>

### ContractValidFeeGranteeMsg
special sudo message to be sent to contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validate_fee_grant` | [ValidateFeeGrant](#archway.gastracker.v1.ValidateFeeGrant) |  |  |






<a name="archway.gastracker.v1.GenesisState"></a>

### GenesisState
Genesis state of the Gastracker module






<a name="archway.gastracker.v1.LeftOverRewardEntry"></a>

### LeftOverRewardEntry
Reward entry per beneficiary address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_rewards` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated |  |






<a name="archway.gastracker.v1.RewardDistributionEvent"></a>

### RewardDistributionEvent
Event emitted when Reward is allocated


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `reward_address` | [string](#string) |  |  |
| `contract_rewards` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
| `leftover_rewards` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated |  |






<a name="archway.gastracker.v1.TransactionTracking"></a>

### TransactionTracking
Tracking contract gas usage and total gas consumption per transaction


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `max_gas_allowed` | [uint64](#uint64) |  |  |
| `max_contract_rewards` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated |  |
| `contract_tracking_infos` | [ContractGasTracking](#archway.gastracker.v1.ContractGasTracking) | repeated |  |
| `is_eligible_for_rewards` | [bool](#bool) |  |  |






<a name="archway.gastracker.v1.ValidateFeeGrant"></a>

### ValidateFeeGrant
special sudo payload to be handled by contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `grantee` | [string](#string) |  |  |
| `gas_fee_to_grant` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
| `msgs` | [WasmMsg](#archway.gastracker.v1.WasmMsg) | repeated |  |






<a name="archway.gastracker.v1.WasmMsg"></a>

### WasmMsg
wasm message sent in a tx


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `msg_type` | [WasmMsgType](#archway.gastracker.v1.WasmMsgType) |  |  |
| `data` | [bytes](#bytes) |  |  |





 <!-- end messages -->


<a name="archway.gastracker.v1.ContractOperation"></a>

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



<a name="archway.gastracker.v1.WasmMsgType"></a>

### WasmMsgType
Wasm message type

| Name | Number | Description |
| ---- | ------ | ----------- |
| WASM_MSG_TYPE_UNSPECIFIED | 0 | Unknown wasm message. It is not used. |
| WASM_MSG_TYPE_EXECUTE | 1 | Execute wasm message |
| WASM_MSG_TYPE_MIGRATE | 2 | Migrate wasm message |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="archway/gastracker/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## archway/gastracker/v1/query.proto



<a name="archway.gastracker.v1.QueryBlockGasTrackingRequest"></a>

### QueryBlockGasTrackingRequest
Request to get the block gas tracking






<a name="archway.gastracker.v1.QueryBlockGasTrackingResponse"></a>

### QueryBlockGasTrackingResponse
Response to get the block gas tracking response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `block_gas_tracking` | [BlockGasTracking](#archway.gastracker.v1.BlockGasTracking) |  |  |






<a name="archway.gastracker.v1.QueryContractMetadataRequest"></a>

### QueryContractMetadataRequest
Request to get contract metadata


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |






<a name="archway.gastracker.v1.QueryContractMetadataResponse"></a>

### QueryContractMetadataResponse
Response to the contract metadata query


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `metadata` | [ContractInstanceMetadata](#archway.gastracker.v1.ContractInstanceMetadata) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="archway.gastracker.v1.Query"></a>

### Query
Query service for Gas tracker

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `ContractMetadata` | [QueryContractMetadataRequest](#archway.gastracker.v1.QueryContractMetadataRequest) | [QueryContractMetadataResponse](#archway.gastracker.v1.QueryContractMetadataResponse) | ContractMetadata returns gastracker metadata of contract | GET|/archway/gastracker/v1/contract/metadata/{address}|
| `BlockGasTracking` | [QueryBlockGasTrackingRequest](#archway.gastracker.v1.QueryBlockGasTrackingRequest) | [QueryBlockGasTrackingResponse](#archway.gastracker.v1.QueryBlockGasTrackingResponse) | BlockGasTracking returns block gas tracking for the latest block | GET|/archway/gastracker/v1/block_gas_tracking|

 <!-- end services -->



<a name="archway/gastracker/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## archway/gastracker/v1/tx.proto



<a name="archway.gastracker.v1.MsgSetContractMetadata"></a>

### MsgSetContractMetadata
Request to set contract metadata


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `contract_address` | [string](#string) |  |  |
| `metadata` | [ContractInstanceMetadata](#archway.gastracker.v1.ContractInstanceMetadata) |  |  |






<a name="archway.gastracker.v1.MsgSetContractMetadataResponse"></a>

### MsgSetContractMetadataResponse
Response returned in rpc call of SetContractMetadata





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="archway.gastracker.v1.Msg"></a>

### Msg
Msg defines the gastracker msg service

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `SetContractMetadata` | [MsgSetContractMetadata](#archway.gastracker.v1.MsgSetContractMetadata) | [MsgSetContractMetadataResponse](#archway.gastracker.v1.MsgSetContractMetadataResponse) | SetContractMetadata to set the gas tracking metadata of contract | |

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

