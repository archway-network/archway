<!--
order: 8
-->

# WASM bindings

Section describes interaction with the module by a contract using CosmWasm custom query and message handling plugins.

## Custom query

[The custom query structure](https://github.com/archway-network/archway/blob/4220b9a643fc37840674a261552f26ec4699a32b/x/rewards/wasmbinding/types/query.go#L12) is used to query the `x/rewards` specific data by a contract.

This query is expected to fail if:

* Query has no request specified (`metadata` and `current_rewards` fields are not defined);
* Query has more than one request specified;

### Metadata

The `metadata` request returns a contract metadata state. A contract can query its own or any other contract's metadata.

Query example:

```json
{
  "metadata": {
    "contract_address": "archway14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sy85n2u"
  }
}
```

Example response:

```json
{
  "owner_address": "archway12reqvcenxgv5s7z96pkytzajtl4lf2epyfman2",
  "rewards_address": "archway12reqvcenxgv5s7z96pkytzajtl4lf2epyfman2"
}
```

### Current rewards

The `current_rewards` request returns the current credited to an account address rewards. A contract can query any account address rewards state.

Query example:

```json
{
  "current_rewards": {
    "rewards_address": "archway12reqvcenxgv5s7z96pkytzajtl4lf2epyfman2"
  }
}
```

Example response:

```json
{
  "rewards": "6460uarch"
}
```

> The `rewards` response field is a string serialized `sdk.Coins` object.

## Custom message

[The custom message structure](https://github.com/archway-network/archway/blob/4220b9a643fc37840674a261552f26ec4699a32b/x/rewards/wasmbinding/types/msg.go#L12) is used to send the `x/rewards` state change messages by a contract.

This message is expected to fail if:

* Message has no sub-message specified (`update_metadata` and `withdraw_rewards` fields are not defined);
* Message has more than one sub-message specified;

### Update metadata

The `update_metadata` request is used to update an existing contract metadata.

Message example (CosmWasm's `CosmosMsg`):

```json
{
  "custom": {
    "update_metadata": {
      "owner_address": "archway14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sy85n2u",
      "rewards_address": "archway14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sy85n2u"
    }
  }
}
```

Sub-message fields:

* `owner_address` - update the contract metadata owner address (optional). Update is skipped if this field is omitted or empty.
* `rewards_address` - update the contract rewards received address (optional). Update is skipped if this field is omitted or empty.

This sub-message doesn't return a response data.

This sub-message is expected to fail if:

* Contract does not exist;
* Metadata is not set for a contract;
* No fields to update were set (`owner_address` and `rewards_address` are empty);
* The contract address is not set as the metadata's `owner_address` (request is unauthorized);

### Withdraw rewards

The `withdraw_rewards` request is used to withdraw the current credited to a contract address reward tokens.

> Contract address is used as the `rewards_address` for this sub-message: a contract can only request withdrawal of funds, credited for his own address.

Message example (CosmWasm's `CosmosMsg`):

```json
{
  "custom": {
    "withdraw_rewards": {}
  }
}
```

Sub-message example response:

```json
{
  "rewards": "6460uarch"
}
```

> The `rewards` response field is a string serialized `sdk.Coins` object.

## Usage examples

### Go contract

The [cosmwasm-go repository](https://github.com/CosmWasm/cosmwasm-go) has the `Voter` example contract. This contract utilizes all the features of the CosmWasm API and the `x/rewards` WASM bindings.

#### Custom message send and reply handling

The [handleMsgWithdrawRewards](https://github.com/CosmWasm/cosmwasm-go/blob/5a075164191c7f55912cbaaca5e0f1ccc5e53348/example/voter/src/handler.go#L362) CosmWasm *Execute* handler sends the custom `withdraw_rewards` message using the WASM bindings.

The [handleReplyCustomWithdrawMsg](https://github.com/CosmWasm/cosmwasm-go/blob/5a075164191c7f55912cbaaca5e0f1ccc5e53348/example/voter/src/handler.go#L390) CosmWasm *Reply* handler parses the `withdraw_rewards` message response.

#### Custom query

The [queryCustomMetadataCustom](https://github.com/CosmWasm/cosmwasm-go/blob/5a075164191c7f55912cbaaca5e0f1ccc5e53348/example/voter/src/querier.go#L192) CosmWasm *Query* handler sends and parses the `metadata` custom query.