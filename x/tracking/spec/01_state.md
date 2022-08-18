<!--
order: 1
-->

# State

Section describes all stored by the module objects and their storage keys.

Refer to the [tracking.proto](../../../proto/archway/tracking/v1beta1/tracking.proto) for objects fields description.

## TxInfo

[TxInfo](../../../proto/archway/tracking/v1beta1/tracking.proto#L22) keeps a transaction gas tracking data.

Example:
```json
{
  "id":1,
  "height": 2,
  "total_gas": 1000
}
```

where: 
* `id` - unique sequentially incremented identificator;
* `height`-  reference to the block height for the transaction;
* `total_gas` - sum of gas consumed by all contract operations (VM + SDK gas);

> TxInfo is created automatically during module EndBlocker.

Storage keys: 
- TxInfo: `0x00 | 0x01 | ID -> ProtocolBuffer(TxInfo)`
- TxInfoByBlock: `0x00 | 0x02 | BlockHeight | ID -> Nil`

## ContractOperationInfo

[ContractOperationInfo](../../../proto/archway/tracking/v1beta1/tracking.proto#L36) keeps a single contract operation gas consumption data.

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
* `id` - unique sequentially incremented identificator;
* `tx_id`-  reference to the [TxInfo](./01_state.md#TxInfo) object;
* `contract_address`-  contract bech32-encoded CosmWasm address;
* `operation_type`-  [enum](../../../proto/archway/tracking/v1beta1/tracking.proto#L9) denoting which operation is consumed gas;
* `vm_gas` - gas consumption reported by the SDK gas meter and the WASM GasRegister (cost of *Execute* / *Query* / etc);
* `sdk_gas` - gas consumption reported by the WASM VM;

Storage keys:
- ContractOperationInfo `0x01 | 0x01 | ID -> ProtocolBuffer(ContractOperationInfo)`
- ContractOperationInfoByTx: `0x01 | 0x02 | TxInfoID | ID -> Nil`
