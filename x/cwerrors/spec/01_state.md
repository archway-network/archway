# State

Section describes all stored by the module objects and their storage keys

## Params

[Params](../../../proto/archway/cwerrors/v1/params.proto) object is used to store the module params

The params value can only be updated by x/gov module via a governance upgrade proposal. [More](./02_messages.md#msgupdateparams)

Storage keys:
* Params: `ParamsKey -> ProtocolBuffer(Params)`

```protobuf
message Params {
    // error_stored_time is the relative block height until which error is stored
    int64 error_stored_time = 1; 
    // subsciption_fee is the fee required to subscribe to error callbacks
    cosmos.base.v1beta1.Coin subscription_fee = 2 [ (gogoproto.nullable) = false ];
    // subscription_period is the period for which the subscription is valid
    int64 subscription_period = 3;
}
```

## ErrorID

ErrorID is a sequence number used to increment error ID.

Storage keys:
* ErrorID: `ErrorIDKey -> uint64`

## Contract Errors

Contract Errors is a collection of all the error ids associated with a given contract address. This is used to query contract errors.

Storage keys:
* ContractErrors: `ContractErrorsKeyPrefix | contractAddress | errorID -> errorID`

## Errors

Errors is a collections of all the [SudoErrors](../../../proto/archway/cwerrors/v1/cwerrors.proto) currently stored by the module which can be queried.

Storage keys:
* Errors: `ErrorsKeyPrefix | errorID -> protobuf(SudoError)`

```protobuf
message SudoError {
    // module_name is the name of the module throwing the error
    string module_name = 1;
    // error_code is the module level error code
    uint32 error_code = 2;
    // contract_address is the address of the contract which will receive the error callback
    string contract_address = 3;
    // input_payload is any input which caused the error
    string input_payload = 4;
    // error_message is the error message
    string error_message = 5;
}
```

## Deletion Blocks

Deletion Blocks is a collection of all the error ids which need to be pruned in a given block height

Storage keys:
* DeletionBlocks: `DeletionBlocksKeyPrefix | blockHeight | errorID -> errorID`

## Contract Subscriptions

Contract Subscriptions is a map of the contract addresses which have subscriptions and the height when the subscription expires

Storage keys:
* Contract Subscriptions: `ContractSubscriptionsKeyPrefix | contractAddress -> deletionHeight`

## Subscription End Block

Subscritption End Block is a collections of all the subscriptions which need to be cleared at the given block height

Storage keys:
* Subscription End Block: `SubscriptionEndBlockKeyPrefix | blockHeight | contractAddress -> contractAddress`

# Transient State

The sudo errors which belong to the contracts with subscription are stored in the transient state of the block.

Transient Storage keys:
* SudoErrors: `ErrorsForSudoCallbackKey | errorId -> SudoError`