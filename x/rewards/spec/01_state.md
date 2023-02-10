<!--
order: 1
-->

# State

Section describes all stored by the module objects and their storage keys.

Refer to the [rewards.proto](../../../proto/archway/rewards/v1beta1/rewards.proto) for objects fields description.

## Params

- Params: `Paramsspace("rewards") -> legacy_amino(params)`

[Params](../../../proto/archway/rewards/v1beta1/rewards.proto#L11) is a module-wide configuration structure.

## Pool

**Rewards** module account is used to aggregate tx fee rebate and inflation rewards tokens and transfer those tokens to a corresponding rewards address during the *withdrawal* operation.

## ContractMetadata

[ContractMetadata](../../../proto/archway/rewards/v1beta1/rewards.proto#L31) object is used to store per contract rewards specific parameters.

Example:

```json
{
  "contract_address": "archway14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sy85n2u",
  "owner_address": "archway12reqvcenxgv5s7z96pkytzajtl4lf2epyfman2",
  "rewards_address": "archway12reqvcenxgv5s7z96pkytzajtl4lf2epyfman2"
}
```

where:

* `contract_address` - contract bech32-encoded CosmWasm address.
* `owner_address` - bech32-encoded contract's owner address.
  * Only the owner is authorized to change the metadata.
  * This field could be an account or a contract address.
  * If it is a contract address, the contract itself could modify the metadata on its own via the WASM bindings functionality.
* `rewards_address` - bech32-encoded account address to receive the contract's rewards via the *withdrawal* operation.

> Contract metadata is not created automatically; it is created by the `MsgSetContractMetadata` transaction which must be signed by a contract admin.
> A contract admin is set by the CosmWasm *Instantiate* operation.

> If the `rewards_address` field is not set (metadata has not been created or the field is empty), a contract won't receive any rewards.

Storage keys:

- ContractMetadata: `0x00 | 0x00 | ContractAddr -> ProtocolBuffer(ContractMetadata)`

## BlockRewards

[BlockRewards](../../../proto/archway/rewards/v1beta1/rewards.proto#L46) object is used to track the inflationary rewards per block that are distributed to dApps in the **BeginBlocker**.

Example:

```json
{
  "height": "100",
  "inflation_rewards": {
    "denom": "uarch",
    "amount": "633764"
  },
  "max_gas": "100000000"
}
```

Entry is created by the [MintBankKeeper](../mintbankkeeper/keeper.go#L25).
This keeper is a wrapper around the standart `x/bank` keeper that transfers tokens between modules and is used by the `x/mint` keeper as a dependency.
Keeper's task is to split minted inflation tokens between the **FeeCollector** (`x/auth`) and the **Rewards** (`x/rewards`) modules using the *InflationRewardsRatio* parameter.

Object is pruned (removed) at the **BeginBlocker**.
Pruning mechanism stores the last 10 entries (last 10 blocks) and a user can query that history.

Storage keys:

* BlockRewards: `0x01 | 0x00 | BlockHeight -> ProtocolBuffer(BlockRewards)`

## TxRewards

[TxRewards](../../../proto/archway/rewards/v1beta1/rewards.proto#L60) object is used to track the tx fee rebate rewards per transaction that are distributed to dApps in the BeginBlocker.

Example:

```json
{
  "tx_id": "10",
  "height": "100",
  "fee_rewards": [
    {
      "denom": "uarch",
      "amount": "6337"
    }
  ]
}
```

Entry is created by the [DeductFeeDecorator](03_ante_handlers.md#DeductFeeDecorator) Ante handler.

The unique entry ID (`tx_id`) is taken from the `x/tracking` module which is the current transaction being processed by the chain.

Object pruning mechanism is the same as the **BlockRewards** one.

Storage keys:

* TxRewards: `0x02 | 0x00 | TxID -> ProtocolBuffer(TxRewards)`
* TxRewardsByBlockHeight:  `0x02 | 0x01 | BlockHeight | TxID -> Nil`

## MinConsensusFee

The *minimum consensus fee* is a price for one transaction gas unit. Value is used to decline transactions with fees lower than the minimum bound.

Value is updated by the **MintBankKeeper** for each block.

This mechanism was introduced to the Archway protocol to avoid cases where one transaction with low fees (or without fees at all) could cause higher dApp rewards, breaking the protocol economic model.

Storage keys:

* MinConsensusFee: `0x03 | 0x00 -> ProtocolBuffer(sdk.Coin)`

## RewardsRecord

[RewardsRecord](../../../proto/archway/rewards/v1beta1/rewards.proto#L78) object is used to track calculated rewards for a particular rewards account within the **BeginBlocker**.
Those records are used later to withdraw calculated rewards for a particular address.

Example:

```json
{
  "id": "1",
  "rewards_address": "archway12reqvcenxgv5s7z96pkytzajtl4lf2epyfman2",
  "rewards": [
    {
      "denom": "uarch",
      "amount": "10000"
    }
  ],
  "calculated_height": 100,
  "calculated_time": {
    "seconds": 1660591975,
    "nanos": 0
  }
}
```

This mechanism was introduced to the Archway protocol to reduce the CPU load on the module's **BeginBlocker** and to give a contract control over its rewards ([WASM bindings section](08_wasm_bindings.md)).

Entries are pruned on a successful *withdrawal* operation.

Storage keys:

* RewardsRecordID: `0x04 | 0x00 -> uint64`
* RewardsRecord: `0x04 | 0x01 | ID -> ProtocolBuffer(RewardsRecord)`
* RewardsRecordByAddress: `0x04 | 0x02 | RewardsAddress | ID -> nil`

## Contract Flat Fees

Flat Fees for a given contract is used as a contract premium which allows smart contract developers to define a custom flat fee for interacting with their smart contract.

Value for a contract can be updated by the contract owner as set in the [ContractMetadata](#contractmetadata). 


Storage keys:

* RewardsRecordByAddress: `0x05 | 0x00 | ContractAddress -> ProtocolBuffer(sdk.Coin)`
