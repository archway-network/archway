<!--
order: 8
-->

# WASM bindings

Section describes interaction with the module by a contract using CosmWasm custom query and message handling plugins.

## Custom query

[The custom query structure](https://github.com/archway-network/archway/blob/4220b9a643fc37840674a261552f26ec4699a32b/x/rewards/wasmbinding/types/query.go#L12) is used to query the `x/rewards` specific data by a contract.

This query is expected to fail if:

* Query has no request specified (`metadata` and `rewards_records` fields are not defined);
* Query has more than one request specified;

### Metadata

The [metadata](https://github.com/archway-network/archway/blob/b027aa56eac2880c03a7bbe85ab9366cd0b59269/x/rewards/wasmbinding/types/query.go#L26) request returns a contract metadata state.
A contract can query its own or any other contract's metadata.

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

### Rewards records

The [rewards_records](https://github.com/archway-network/archway/blob/b027aa56eac2880c03a7bbe85ab9366cd0b59269/x/rewards/wasmbinding/types/query.go#L42) request returns the paginated list of `RewardsRecord` objects credited to an account address.
A [RewardsRecord](https://github.com/archway-network/archway/blob/b027aa56eac2880c03a7bbe85ab9366cd0b59269/x/rewards/wasmbinding/types/query.go#L59) entry contains a portion of credited rewards by a specific contract at a block height.
A contract can query any account address rewards state.

This query is paginated to limit the size of the response.
Refer to the [PageRequest](https://github.com/archway-network/archway/blob/b027aa56eac2880c03a7bbe85ab9366cd0b59269/x/rewards/wasmbinding/types/pagination.go#L8) structure description to learn more about the pagination options.
The query response contains the [PageResponse](https://github.com/archway-network/archway/blob/b027aa56eac2880c03a7bbe85ab9366cd0b59269/x/rewards/wasmbinding/types/pagination.go#L28) structure that should be used to query the next page.

> The maximum page limit is bounded by the `MaxWithdrawRecords` parameter.
> 
> If the page limit is not set, the default value is `MaxWithdrawRecords`.

Query example:

```json
{
  "rewards_records": {
    "rewards_address": "archway1allzevxuve88s75pjmcupxhy95qrvjlgvjtf0n",
    "pagination": {
      "limit": 100
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

## Custom message

[The custom message structure](https://github.com/archway-network/archway/blob/4220b9a643fc37840674a261552f26ec4699a32b/x/rewards/wasmbinding/types/msg.go#L12) is used to send the `x/rewards` state change messages by a contract.

This message is expected to fail if:

* Message has no sub-message specified (`update_metadata` and `withdraw_rewards` fields are not defined);
* Message has more than one sub-message specified;

### Update metadata

The [update_metadata](https://github.com/archway-network/archway/blob/b027aa56eac2880c03a7bbe85ab9366cd0b59269/x/rewards/wasmbinding/types/msg.go#L24) request is used to update an existing contract metadata.

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

The [withdraw_rewards](https://github.com/archway-network/archway/blob/b027aa56eac2880c03a7bbe85ab9366cd0b59269/x/rewards/wasmbinding/types/msg.go#L34) request is used to withdraw the current credited to a contract address reward tokens.

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
    "withdraw_rewards": {
      "records_limit": 100
    }
  }
}
```

Sub-message returns the [response](https://github.com/archway-network/archway/blob/b027aa56eac2880c03a7bbe85ab9366cd0b59269/x/rewards/wasmbinding/types/msg.go#L45) that can be handled with the *Reply* CosmWasm functionality.

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

## Usage examples

### Go contract

The [cosmwasm-go repository](https://github.com/CosmWasm/cosmwasm-go) has the `Voter` example contract. This contract utilizes all the features of the CosmWasm API and the `x/rewards` WASM bindings.

#### Custom message send and reply handling

The [handleMsgWithdrawRewards](https://github.com/CosmWasm/cosmwasm-go/blob/45b9f015c12e75f12c0bb4b9c2a27da606a58f4e/example/voter/src/handler.go#L362) CosmWasm *Execute* handler sends the custom `withdraw_rewards` message using the WASM bindings.

The [handleReplyCustomWithdrawMsg](https://github.com/CosmWasm/cosmwasm-go/blob/45b9f015c12e75f12c0bb4b9c2a27da606a58f4e/example/voter/src/handler.go#L390) CosmWasm *Reply* handler parses the `withdraw_rewards` message response.

#### Custom query

The [queryCustomMetadataCustom](https://github.com/CosmWasm/cosmwasm-go/blob/45b9f015c12e75f12c0bb4b9c2a27da606a58f4e/example/voter/src/querier.go#L192) CosmWasm *Query* handler sends and parses the `metadata` custom query.
