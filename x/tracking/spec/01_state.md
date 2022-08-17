<!--
order: 1
-->

# State

Section describes all stored by the module objects and their storage keys.

Refer to the [tracking.proto](../../../proto/archway/tracking/v1beta1/tracking.proto) for objects fields description.

## TxInfo

[TxInfo](https://github.com/archway-network/archway/blob/e130d74bd456be037b4e60dea7dada5d7a8760b5/proto/archway/tracking/v1beta1/tracking.proto#L22) Keeps transaction gas tracking data.

Example:
```json
{
  "id":1,
  "height": 2,
  "total_gas": 1000
}
```

where: 
* `id` - unique sequentially incremented identificator.
* `height`-  reference to the block height for the transaction.
* `total_gas` - sum of gas consumed by all contract operations (VM + SDK gas).

> TxInfo is created automatically during module Endblocker.

Storage keys: 
- TxInfo: `0x00 | 0x01 | ID -> ProtocolBuffer(TxInfo)`
- TxInfoBlock: `0x00 | 0x02 | height | ID -> Nil`

## ContractOperationInfo

[ContractOperationInfo](https://github.com/archway-network/archway/blob/e130d74bd456be037b4e60dea7dada5d7a8760b5/proto/archway/tracking/v1beta1/tracking.proto#L36) keeps single contract operation gas consumption data.

```json
{
  "id":1,
  "tx_id": 2,
  "contract_address": "archway14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sy85n2u",
  "operation_type": 1,
  "vm_gas": 500,
  "sdk_gas": 500
}
```

where: 
* `id` - unique sequentially incremented identificator.
* `tx_id`-  reference to `tx_id` from [TxInfo](./01_state.md#TxInfo).
* `contract_address`-  contract bech32-encoded CosmWasm address. 
* `operation_type`-  enum denoting which operation is consumed gas, [ContractOperation](https://github.com/archway-network/archway/blob/e130d74bd456be037b4e60dea7dada5d7a8760b5/proto/archway/tracking/v1beta1/tracking.proto#L9)
* `vm_gas` - gas consumption reported by the SDK gas meter and the WASM GasRegister (cost of Execute/Query/etc).
* `sdk_gas` - gas consumption reported by the WASM VM.

Storage keys:
- ContractOperationInfo `0x01 | 0x01 | ID -> ProtocolBuffer(ContractOperationInfo)`
- ContractOperationInfoTxIndex: `0x01 | 0x02 | TxInfoId | ID -> Nil`



