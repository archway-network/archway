<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [archway/gastracker/v1/types.proto](#archway/gastracker/v1/types.proto)
    - [BlockGasTracking](#archway.gastracker.v1.BlockGasTracking)
    - [ContractGasTracking](#archway.gastracker.v1.ContractGasTracking)
    - [ContractInstanceMetadata](#archway.gastracker.v1.ContractInstanceMetadata)
    - [ContractRewardCalculationEvent](#archway.gastracker.v1.ContractRewardCalculationEvent)
    - [GenesisState](#archway.gastracker.v1.GenesisState)
    - [LeftOverRewardEntry](#archway.gastracker.v1.LeftOverRewardEntry)
    - [Params](#archway.gastracker.v1.Params)
    - [RewardDistributionEvent](#archway.gastracker.v1.RewardDistributionEvent)
    - [TransactionTracking](#archway.gastracker.v1.TransactionTracking)
  
    - [ContractOperation](#archway.gastracker.v1.ContractOperation)
  
- [archway/gastracker/v1/query.proto](#archway/gastracker/v1/query.proto)
    - [QueryBlockGasTrackingRequest](#archway.gastracker.v1.QueryBlockGasTrackingRequest)
    - [QueryBlockGasTrackingResponse](#archway.gastracker.v1.QueryBlockGasTrackingResponse)
    - [QueryContractMetadataRequest](#archway.gastracker.v1.QueryContractMetadataRequest)
    - [QueryContractMetadataResponse](#archway.gastracker.v1.QueryContractMetadataResponse)
    - [QueryParamsRequest](#archway.gastracker.v1.QueryParamsRequest)
    - [QueryParamsResponse](#archway.gastracker.v1.QueryParamsResponse)
  
    - [Query](#archway.gastracker.v1.Query)
  
- [archway/gastracker/v1/tx.proto](#archway/gastracker/v1/tx.proto)
    - [MsgSetContractMetadata](#archway.gastracker.v1.MsgSetContractMetadata)
    - [MsgSetContractMetadataResponse](#archway.gastracker.v1.MsgSetContractMetadataResponse)
  
    - [Msg](#archway.gastracker.v1.Msg)
  
- [archway/rewards/v1beta1/rewards.proto](#archway/rewards/v1beta1/rewards.proto)
    - [BlockRewards](#archway.rewards.v1beta1.BlockRewards)
    - [BlockTracking](#archway.rewards.v1beta1.BlockTracking)
    - [ContractMetadata](#archway.rewards.v1beta1.ContractMetadata)
    - [Params](#archway.rewards.v1beta1.Params)
    - [TxRewards](#archway.rewards.v1beta1.TxRewards)
  
- [archway/rewards/v1beta1/events.proto](#archway/rewards/v1beta1/events.proto)
    - [ContractMetadataSetEvent](#archway.rewards.v1beta1.ContractMetadataSetEvent)
    - [ContractRewardCalculationEvent](#archway.rewards.v1beta1.ContractRewardCalculationEvent)
    - [ContractRewardDistributionEvent](#archway.rewards.v1beta1.ContractRewardDistributionEvent)
  
- [archway/rewards/v1beta1/genesis.proto](#archway/rewards/v1beta1/genesis.proto)
    - [GenesisState](#archway.rewards.v1beta1.GenesisState)
  
- [archway/rewards/v1beta1/query.proto](#archway/rewards/v1beta1/query.proto)
    - [QueryBlockRewardsTrackingRequest](#archway.rewards.v1beta1.QueryBlockRewardsTrackingRequest)
    - [QueryBlockRewardsTrackingResponse](#archway.rewards.v1beta1.QueryBlockRewardsTrackingResponse)
    - [QueryContractMetadataRequest](#archway.rewards.v1beta1.QueryContractMetadataRequest)
    - [QueryContractMetadataResponse](#archway.rewards.v1beta1.QueryContractMetadataResponse)
    - [QueryParamsRequest](#archway.rewards.v1beta1.QueryParamsRequest)
    - [QueryParamsResponse](#archway.rewards.v1beta1.QueryParamsResponse)
    - [QueryRewardsPoolRequest](#archway.rewards.v1beta1.QueryRewardsPoolRequest)
    - [QueryRewardsPoolResponse](#archway.rewards.v1beta1.QueryRewardsPoolResponse)
  
    - [Query](#archway.rewards.v1beta1.Query)
  
- [archway/rewards/v1beta1/tx.proto](#archway/rewards/v1beta1/tx.proto)
    - [MsgSetContractMetadata](#archway.rewards.v1beta1.MsgSetContractMetadata)
    - [MsgSetContractMetadataResponse](#archway.rewards.v1beta1.MsgSetContractMetadataResponse)
  
    - [Msg](#archway.rewards.v1beta1.Msg)
  
- [archway/tracking/v1beta1/tracking.proto](#archway/tracking/v1beta1/tracking.proto)
    - [BlockTracking](#archway.tracking.v1beta1.BlockTracking)
    - [ContractOperationInfo](#archway.tracking.v1beta1.ContractOperationInfo)
    - [TxInfo](#archway.tracking.v1beta1.TxInfo)
    - [TxTracking](#archway.tracking.v1beta1.TxTracking)
  
    - [ContractOperation](#archway.tracking.v1beta1.ContractOperation)
  
- [archway/tracking/v1beta1/genesis.proto](#archway/tracking/v1beta1/genesis.proto)
    - [GenesisState](#archway.tracking.v1beta1.GenesisState)
  
- [archway/tracking/v1beta1/query.proto](#archway/tracking/v1beta1/query.proto)
    - [QueryBlockGasTrackingRequest](#archway.tracking.v1beta1.QueryBlockGasTrackingRequest)
    - [QueryBlockGasTrackingResponse](#archway.tracking.v1beta1.QueryBlockGasTrackingResponse)
  
    - [Query](#archway.tracking.v1beta1.Query)
  
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






<a name="archway.gastracker.v1.ContractRewardCalculationEvent"></a>

### ContractRewardCalculationEvent
Event emitted when contract reward is calculated


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  |  |
| `gas_consumed` | [uint64](#uint64) |  |  |
| `inflation_rewards` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  |  |
| `contract_rewards` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated |  |
| `metadata` | [ContractInstanceMetadata](#archway.gastracker.v1.ContractInstanceMetadata) |  |  |






<a name="archway.gastracker.v1.GenesisState"></a>

### GenesisState
Genesis state of the Gastracker module






<a name="archway.gastracker.v1.LeftOverRewardEntry"></a>

### LeftOverRewardEntry
Reward entry per beneficiary address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_rewards` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated |  |






<a name="archway.gastracker.v1.Params"></a>

### Params



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gas_tracking_switch` | [bool](#bool) |  |  |
| `gas_rebate_to_user_switch` | [bool](#bool) |  |  |
| `contract_premium_switch` | [bool](#bool) |  |  |
| `dapp_inflation_rewards_ratio` | [string](#string) |  |  |
| `dapp_tx_fee_rebate_ratio` | [string](#string) |  |  |






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






<a name="archway.gastracker.v1.QueryParamsRequest"></a>

### QueryParamsRequest







<a name="archway.gastracker.v1.QueryParamsResponse"></a>

### QueryParamsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#archway.gastracker.v1.Params) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="archway.gastracker.v1.Query"></a>

### Query
Query service for Gas tracker

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#archway.gastracker.v1.QueryParamsRequest) | [QueryParamsResponse](#archway.gastracker.v1.QueryParamsResponse) | Params returns parameters. | GET|/archway/gastracker/v1/params|
| `ContractMetadata` | [QueryContractMetadataRequest](#archway.gastracker.v1.QueryContractMetadataRequest) | [QueryContractMetadataResponse](#archway.gastracker.v1.QueryContractMetadataResponse) | ContractMetadata returns gastracker metadata of contract | GET|/archway/gastracker/v1/contract_metadata/{address}|
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



<a name="archway/rewards/v1beta1/rewards.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## archway/rewards/v1beta1/rewards.proto



<a name="archway.rewards.v1beta1.BlockRewards"></a>

### BlockRewards
BlockRewards defines block related rewards distribution data.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `height` | [int64](#int64) |  | height defines the block height. |
| `inflation_rewards` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | inflation_rewards is the rewards to be distributed. |
| `max_gas` | [uint64](#uint64) |  | max_gas defines the maximum gas for the block that is used to distribute inflation rewards (consensus parameter). |






<a name="archway.rewards.v1beta1.BlockTracking"></a>

### BlockTracking
BlockTracking is the tracking information for a block.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `inflation_rewards` | [BlockRewards](#archway.rewards.v1beta1.BlockRewards) |  | inflation_rewards defines the inflation rewards for the block. |
| `tx_rewards` | [TxRewards](#archway.rewards.v1beta1.TxRewards) | repeated | tx_rewards defines the transaction rewards for the block. |






<a name="archway.rewards.v1beta1.ContractMetadata"></a>

### ContractMetadata
ContractMetadata defines the contract rewards distribution options for a particular contract.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  | contract_address defines the contract address (bech32 encoded). |
| `owner_address` | [string](#string) |  | owner_address is the contract owner address that can modify contract reward options (bech32 encoded). That could be the contract admin or the contract itself. If owner_address is set to contract address, contract can modify the metadata on its own using WASM bindings. |
| `rewards_address` | [string](#string) |  | rewards_address is an address to distribute rewards to (bech32 encoded). If not set (empty), rewards are not distributed for this contract. |






<a name="archway.rewards.v1beta1.Params"></a>

### Params
Params defines the module parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `inflation_rewards_ratio` | [string](#string) |  | inflation_rewards_ratio defines the percentage of minted inflation tokens that are used for dApp rewards [0.0, 1.0]. If set to 0.0, no inflation rewards are distributed. |
| `tx_fee_rebate_ratio` | [string](#string) |  | tx_fee_rebate_ratio defines the percentage of tx fees that are used for dApp rewards [0.0, 1.0]. If set to 0.0, no fee rewards are distributed. |






<a name="archway.rewards.v1beta1.TxRewards"></a>

### TxRewards
TxRewards defines transaction related rewards distribution data.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_id` | [uint64](#uint64) |  | tx_id is the tracking transaction ID (x/tracking is the data source for this value). |
| `height` | [int64](#int64) |  | height defines the block height. |
| `fee_rewards` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | fee_rewards is the rewards to be distributed. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="archway/rewards/v1beta1/events.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## archway/rewards/v1beta1/events.proto



<a name="archway.rewards.v1beta1.ContractMetadataSetEvent"></a>

### ContractMetadataSetEvent
ContractMetadataSetEvent is emitted when the contract metadata is created or updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  | contract_address defines the contract address. |
| `metadata` | [ContractMetadata](#archway.rewards.v1beta1.ContractMetadata) |  | metadata defines the new contract metadata state. |






<a name="archway.rewards.v1beta1.ContractRewardCalculationEvent"></a>

### ContractRewardCalculationEvent
ContractRewardCalculationEvent is emitted when the contract reward is calculated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  | contract_address defines the contract address. |
| `gas_consumed` | [uint64](#uint64) |  | gas_consumed defines the total gas consumption by all WASM operations within one transaction. |
| `inflation_rewards` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | inflation_rewards defines the inflation rewards portions of the rewards. |
| `fee_rebate_rewards` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | fee_rebate_rewards defines the fee rebate rewards portions of the rewards. |
| `metadata` | [ContractMetadata](#archway.rewards.v1beta1.ContractMetadata) |  | metadata defines the contract metadata (if set). |






<a name="archway.rewards.v1beta1.ContractRewardDistributionEvent"></a>

### ContractRewardDistributionEvent
ContractRewardDistributionEvent is emitted when the contract reward is distributed to the corresponding rewards address.
This event might not follow the ContractRewardCalculationEvent if the contract has no metadata set or rewards address is empty.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  | contract_address defines the contract address. |
| `reward_address` | [string](#string) |  | rewards_address defines the rewards address rewards are distributed to. |
| `rewards` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | rewards defines the total rewards being distributed. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="archway/rewards/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## archway/rewards/v1beta1/genesis.proto



<a name="archway.rewards.v1beta1.GenesisState"></a>

### GenesisState
GenesisState defines the initial state of the tracking module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#archway.rewards.v1beta1.Params) |  | params defines all the module parameters. |
| `contracts_metadata` | [ContractMetadata](#archway.rewards.v1beta1.ContractMetadata) | repeated | contracts_metadata defines a list of all contracts metadata. |
| `block_rewards` | [BlockRewards](#archway.rewards.v1beta1.BlockRewards) | repeated | block_rewards defines a list of all block rewards objects. |
| `tx_rewards` | [TxRewards](#archway.rewards.v1beta1.TxRewards) | repeated | tx_rewards defines a list of all tx rewards objects. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="archway/rewards/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## archway/rewards/v1beta1/query.proto



<a name="archway.rewards.v1beta1.QueryBlockRewardsTrackingRequest"></a>

### QueryBlockRewardsTrackingRequest
QueryBlockRewardsTrackingRequest is the request for Query.BlockRewardsTracking.






<a name="archway.rewards.v1beta1.QueryBlockRewardsTrackingResponse"></a>

### QueryBlockRewardsTrackingResponse
QueryBlockRewardsTrackingResponse is the response for Query.BlockRewardsTracking.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `block` | [BlockTracking](#archway.rewards.v1beta1.BlockTracking) |  |  |






<a name="archway.rewards.v1beta1.QueryContractMetadataRequest"></a>

### QueryContractMetadataRequest
QueryContractMetadataRequest is the request for Query.ContractMetadata.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  | contract_address is the contract address (bech32 encoded). |






<a name="archway.rewards.v1beta1.QueryContractMetadataResponse"></a>

### QueryContractMetadataResponse
QueryContractMetadataResponse is the response for Query.ContractMetadata.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `metadata` | [ContractMetadata](#archway.rewards.v1beta1.ContractMetadata) |  |  |






<a name="archway.rewards.v1beta1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request for Query.Params.






<a name="archway.rewards.v1beta1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response for Query.Params.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#archway.rewards.v1beta1.Params) |  |  |






<a name="archway.rewards.v1beta1.QueryRewardsPoolRequest"></a>

### QueryRewardsPoolRequest
QueryRewardsPoolRequest is the request for Query.RewardsPool.






<a name="archway.rewards.v1beta1.QueryRewardsPoolResponse"></a>

### QueryRewardsPoolResponse
QueryRewardsPoolResponse is the response for Query.RewardsPool.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `funds` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="archway.rewards.v1beta1.Query"></a>

### Query
Query service for the tracking module.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#archway.rewards.v1beta1.QueryParamsRequest) | [QueryParamsResponse](#archway.rewards.v1beta1.QueryParamsResponse) | Params returns module parameters. | GET|/archway/rewards/v1/params|
| `ContractMetadata` | [QueryContractMetadataRequest](#archway.rewards.v1beta1.QueryContractMetadataRequest) | [QueryContractMetadataResponse](#archway.rewards.v1beta1.QueryContractMetadataResponse) | ContractMetadata returns the contract rewards parameters (metadata). | GET|/archway/rewards/v1/contract_metadata|
| `BlockRewardsTracking` | [QueryBlockRewardsTrackingRequest](#archway.rewards.v1beta1.QueryBlockRewardsTrackingRequest) | [QueryBlockRewardsTrackingResponse](#archway.rewards.v1beta1.QueryBlockRewardsTrackingResponse) | BlockRewardsTracking returns block rewards tracking for the current block. | GET|/archway/rewards/v1/block_rewards_tracking|
| `RewardsPool` | [QueryRewardsPoolRequest](#archway.rewards.v1beta1.QueryRewardsPoolRequest) | [QueryRewardsPoolResponse](#archway.rewards.v1beta1.QueryRewardsPoolResponse) | RewardsPool returns the current undistributed rewards pool funds. | GET|/archway/rewards/v1/rewards_pool|

 <!-- end services -->



<a name="archway/rewards/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## archway/rewards/v1beta1/tx.proto



<a name="archway.rewards.v1beta1.MsgSetContractMetadata"></a>

### MsgSetContractMetadata
MsgSetContractMetadata is the request for Msg.SetContractMetadata.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender_address` | [string](#string) |  | sender_address is the msg sender address (bech32 encoded). |
| `metadata` | [ContractMetadata](#archway.rewards.v1beta1.ContractMetadata) |  | metadata is the contract metadata to set / update. If metadata exists, non-empty fields will be updated. |






<a name="archway.rewards.v1beta1.MsgSetContractMetadataResponse"></a>

### MsgSetContractMetadataResponse
MsgSetContractMetadataResponse is the response for Msg.SetContractMetadata.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="archway.rewards.v1beta1.Msg"></a>

### Msg
Msg defines the module messaging service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `SetContractMetadata` | [MsgSetContractMetadata](#archway.rewards.v1beta1.MsgSetContractMetadata) | [MsgSetContractMetadataResponse](#archway.rewards.v1beta1.MsgSetContractMetadataResponse) | SetContractMetadata creates or updates an existing contract metadata. Method is authorized to the contract owner (admin if no metadata exists). | |

 <!-- end services -->



<a name="archway/tracking/v1beta1/tracking.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## archway/tracking/v1beta1/tracking.proto



<a name="archway.tracking.v1beta1.BlockTracking"></a>

### BlockTracking
BlockTracking is the tracking information for a block.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `txs` | [TxTracking](#archway.tracking.v1beta1.TxTracking) | repeated | txs defines the list of transactions tracked in the block. |






<a name="archway.tracking.v1beta1.ContractOperationInfo"></a>

### ContractOperationInfo
ContractOperationInfo keeps a single contract operation gas consumption data.
Object is being created by the IngestGasRecord call from the wasmd.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  | id defines the unique operation ID. |
| `tx_id` | [uint64](#uint64) |  | tx_id defines a transaction ID operation relates to (TxInfo.id). |
| `contract_address` | [string](#string) |  | contract_address defines the contract address operation relates to. |
| `operation_type` | [ContractOperation](#archway.tracking.v1beta1.ContractOperation) |  | operation_type defines the gas consumption type. |
| `vm_gas` | [uint64](#uint64) |  | vm_gas is the gas consumption reported by the WASM VM. Value is adjusted by this module (CalculateUpdatedGas func). |
| `sdk_gas` | [uint64](#uint64) |  | sdk_gas is the gas consumption reported by the SDK gas meter and the WASM GasRegister (cost of Execute/Query/etc). Value is adjusted by this module (CalculateUpdatedGas func). |






<a name="archway.tracking.v1beta1.TxInfo"></a>

### TxInfo
TxInfo keeps a transaction gas tracking data.
Object is being created at the module EndBlocker.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  | id defines the unique transaction ID. |
| `height` | [int64](#int64) |  | height defines the block height of the transaction. |
| `total_gas` | [uint64](#uint64) |  | total_gas defines total gas consumption by the transaction. It is the sum of gas consumed by all contract operations (VM + SDK gas). |






<a name="archway.tracking.v1beta1.TxTracking"></a>

### TxTracking
TxTracking is the tracking information for a single transaction.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `info` | [TxInfo](#archway.tracking.v1beta1.TxInfo) |  | info defines the transaction details. |
| `contract_operations` | [ContractOperationInfo](#archway.tracking.v1beta1.ContractOperationInfo) | repeated | contract_operations defines the list of contract operations consumed by the transaction. |





 <!-- end messages -->


<a name="archway.tracking.v1beta1.ContractOperation"></a>

### ContractOperation
ContractOperation denotes which operation consumed gas.

| Name | Number | Description |
| ---- | ------ | ----------- |
| CONTRACT_OPERATION_UNSPECIFIED | 0 | Invalid or unknown operation |
| CONTRACT_OPERATION_INSTANTIATION | 1 | Instantiate operation |
| CONTRACT_OPERATION_EXECUTION | 2 | Execute operation |
| CONTRACT_OPERATION_QUERY | 3 | Query |
| CONTRACT_OPERATION_MIGRATE | 4 | Migrate operation |
| CONTRACT_OPERATION_IBC | 5 | IBC operations |
| CONTRACT_OPERATION_SUDO | 6 | Sudo operation |
| CONTRACT_OPERATION_REPLY | 7 | Reply callback operation |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="archway/tracking/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## archway/tracking/v1beta1/genesis.proto



<a name="archway.tracking.v1beta1.GenesisState"></a>

### GenesisState
GenesisState defines the initial state of the tracking module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_infos` | [TxInfo](#archway.tracking.v1beta1.TxInfo) | repeated | tx_infos defines a list of all the tracked transactions. |
| `contract_op_infos` | [ContractOperationInfo](#archway.tracking.v1beta1.ContractOperationInfo) | repeated | contract_op_infos defines a list of all the tracked contract operations. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="archway/tracking/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## archway/tracking/v1beta1/query.proto



<a name="archway.tracking.v1beta1.QueryBlockGasTrackingRequest"></a>

### QueryBlockGasTrackingRequest
QueryBlockGasTrackingRequest is the request for Query.BlockGasTracking.






<a name="archway.tracking.v1beta1.QueryBlockGasTrackingResponse"></a>

### QueryBlockGasTrackingResponse
QueryBlockGasTrackingResponse is the response for Query.BlockGasTracking.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `block` | [BlockTracking](#archway.tracking.v1beta1.BlockTracking) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="archway.tracking.v1beta1.Query"></a>

### Query
Query service for the tracking module.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `BlockGasTracking` | [QueryBlockGasTrackingRequest](#archway.tracking.v1beta1.QueryBlockGasTrackingRequest) | [QueryBlockGasTrackingResponse](#archway.tracking.v1beta1.QueryBlockGasTrackingResponse) | BlockGasTracking returns block gas tracking for the current block | GET|/archway/tracking/v1/block_gas_tracking|

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

