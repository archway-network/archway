<!--
order: 8
-->

# WASM bindings

Section describes interaction with the module by a contract using CosmWasm custom query and message handling plugins.

## Custom query

[The custom query structure](../../../wasmbinding/types/query.go#L10) is used to query a specific module data by a contract.

This query is expected to fail if:

* Query has no sub-query specified (`rewards` field is not defined);
* Query has more than one sub-query specified;

## Custom message

[The custom message structure](../../../wasmbinding/types/msg.go#L10) is used to send a module specific state change message by a contract.

This message is expected to fail if:

* Message has no sub-message specified (`rewards` field is not defined);
* Message has more than one sub-message specified;

## Rewards module bindings

### Queries

[The sub-query structure](../../../wasmbinding/rewards/types/query.go#L8) is used to query the `x/rewards` module specific data.

This query is expected to fail if:

* Query has no request specified (`metadata` and `rewards_records` fields are not defined);
* Query has more than one request specified;

#### Metadata

The [metadata](../../../wasmbinding/rewards/types/query_metadata.go#L12) request returns a contract metadata state.
A contract can query its own or any other contract's metadata.

Query example:

```json
{
  "rewards": {
    "metadata": {
      "contract_address": "archway14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sy85n2u"
    }
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

#### Rewards records

The [rewards_records](../../../wasmbinding/rewards/types/query_records.go#L17) request returns the paginated list of `RewardsRecord` objects credited to an account address.
A [RewardsRecord](../../../wasmbinding/rewards/types/query_records.go#L35) entry contains a portion of credited rewards by a specific contract at a block height.
A contract can query any account address rewards state.

This query is paginated to limit the size of the response.
Refer to the [PageRequest](../../../wasmbinding/pkg/pagination.go#L8) structure description to learn more about the pagination options.
The query response contains the [PageResponse](../../../wasmbinding/pkg/pagination.go#L28) structure that should be used to query the next page.

> The maximum page limit is bounded by the `MaxWithdrawRecords` parameter.
>
> If the page limit is not set, the default value is `MaxWithdrawRecords`.

Query example:

```json
{
  "rewards": {
    "rewards_records": {
      "rewards_address": "archway1allzevxuve88s75pjmcupxhy95qrvjlgvjtf0n",
      "pagination": {
        "limit": 100
      }
    }
  }
}
```

Example response:

```json
{
  "records": [
    {
      "id": 3,
      "rewards_address": "archway1allzevxuve88s75pjmcupxhy95qrvjlgvjtf0n",
      "rewards": [
        {
          "amount": "6463",
          "denom": "uarch"
        }
      ],
      "calculated_height": 38,
      "calculated_time": "2022-08-17T05:07:35.462087Z"
    }
  ],
  "pagination": {
    "total": 200
  }
}
```

#### FlatFee

The [flatfee](../../../wasmbinding/rewards/types/query_flatfee.go) request returns a contract flat fee.
A contract can query its own or any other contract's metadata.

Query example:

```json
{
  "rewards": {
    "flat_fee": {
      "contract_address": "archway14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sy85n2u"
    }
  }
}
```

Example response:

```json
{
  "flat_fee_amount": {
    "amount": "10000",
    "denom": "uarch"
  }
}
```

### Messages

[The sub-message structure](../../../wasmbinding/rewards/types/msg.go#L8) is used to send the `x/rewards` module specific state change message.

This message is expected to fail if:

* Message has no operations specified (`update_metadata` and `withdraw_rewards` and `set_flat_fee` fields are not defined);
* Message has more than one operation specified;

#### Update metadata

The [update_metadata](../../../wasmbinding/rewards/types/msg_metadata.go#L12) request is used to update an existing contract metadata.

Message example (CosmWasm's `CosmosMsg`):

```json
{
  "custom": {
    "rewards": {
      "update_metadata": {
        "owner_address": "archway14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sy85n2u",
        "rewards_address": "archway14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sy85n2u"
      }
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

#### Withdraw rewards

The [withdraw_rewards](../../../wasmbinding/rewards/types/msg_withdraw.go#L12) request is used to withdraw the current credited to a contract address reward tokens.

> Contract address is used as the `rewards_address` for this sub-message: a contract can only request withdrawal of funds, credited for his own address.

This sub-message uses `RewardsRecord` objects that are created for a specific `rewards_address` during the dApp rewards distribution.
The `withdraw-rewards` command has two operation modes, which defines which `RewardsRecord` objects to process:

* *Records by limit* - select the first N `RewardsRecord` objects available;
* *Records by IDs* - select specific `RewardsRecord` objects by their IDs;

Sub-message is expected to fail if:

* Specified `records_limit` field value or the length of `record_ids` exceeds the `MaxWithdrawRecords` module parameter;
* The `records_limit` and the `record_ids` fields are both set (one of is allowed);
* Provided record ID is not found;
* Provided record ID is not linked to the `contract_address`;

Message example (CosmWasm's `CosmosMsg`):

```json
{
  "custom": {
    "rewards": {
      "withdraw_rewards": {
        "records_limit": 100
      }
    }
  }
}
```

Sub-message returns the [response](../../../wasmbinding/rewards/types/msg_withdraw.go#L23) that can be handled with the *Reply* CosmWasm functionality.

Response example:

```json
{
  "records_num": 100,
  "total_rewards": [
    {
      "amount": "6463",
      "denom": "uarch"
    }
  ]
}
```


#### Set Flat Fee

The [set_flat_fee](../../../wasmbinding/rewards/types/msg_flatfee.go#L12) request is used to update an existing contract metadata.

Message example (CosmWasm's `CosmosMsg`):

```json
{
  "custom": {
    "rewards": {
      "set_flat_fee": {
        "contract_address": "archway14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sy85n2u",
        "flat_fee_amount": {
          "amount": "10000",
          "denom": "uarch"
        }
      }
    }
  }
}
```

Sub-message fields:

* `contract_address` - the contract address to update the flat fee for.
* `flat_fee_amount` - flat fee amount .

This sub-message doesn't return a response data.
If the message coins zero value amount, the existing flat fee is removed.

This sub-message is expected to fail if:

* Contract does not exist;
* Metadata is not set for a contract;
* The contract address is not set as the metadata's `owner_address` (request is unauthorized);

## Usage examples

### Go contract

This repository has the [Voter](../../../contracts/go/voter) example contract.
This contract utilizes all the features of the CosmWasm API and the `x/rewards` WASM bindings.

#### Custom message send and reply handling

The [handleMsgWithdrawRewards](../../../contracts/go/voter/src/handler.go#L378) CosmWasm *Execute* handler sends the custom `withdraw_rewards` message using the WASM bindings.

The [handleReplyCustomWithdrawMsg](../../../contracts/go/voter/src/handler.go#L408) CosmWasm *Reply* handler parses the `withdraw_rewards` message response.

#### Custom query

The [queryCustomMetadataCustom](../../../contracts/go/voter/src/querier.go#L213) CosmWasm *Query* handler sends and parses the `metadata` custom query.
