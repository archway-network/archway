# ADR-012 – CW Errors module

Date: 2023-04-01

## Status

Implemented

## Abstract

We propose a x/cw-errors module which enables a standardized way for the CosmWasm smart contracts to receive errors regarding their executions which are initiated by protocol.

This is useful for contracts who use the features provided by x/callback anc x/cwica modules.

## Context

With the introduction of x/callback and x/cwica modules, there are multiple mechanisms where the protocol executes a smart contract. When a user executes a contract and there is an error, the user receives the error and can act on it. However, when the protocol executes a contract and there is an error, the error could get lost in the node logs and there is no way for the contract or the developer to become aware of the error neing thrown at all.

### x/callback

When the protocol executes a callback, there could be error thrown due to
1. The contract does not have the expected sudo entrypoint
2. The callback execution might consume more gas than is allowed
3. There could be an error thrown from the contract

### x/cwica

When a contract wants to execute an ica tx on another chain, there could be error thrown due to 
1. The counterparty could not unmarshall the msg, or did not recognize the msg
2. The counterparty tx execution failed
3. The ibc packet timedout

Under the above mentioned situations, we need a way to convey the error and its details to the contract. 

## Decision

We will implement a new module called x/cwerrors which allows the protocol invoked contract errors to be able know an error has occurred.

When a module which does Sudo execution on a contract encounters an error, it will execute the `cwErrorsKeeper.SetError(ctx, sudoErr)` where the `sudoErr` is of the following type

```protobuf
message SudoError {
  // module_name is the name of the module throwing the error
  string module_name = 1;
  // error_code is the module level error code. Every module needs to implement an error id for unique errors they want to throw
  int32 error_code = 2;
  // contract_address is the address of the contract which will receive the error callback
  string contract_address = 3;
  // input_payload is any input which caused the error
  string input_payload = 4;
  // error_message is the error message
  string error_message = 5;
}
```
When the x/cwerror module receives a contract error (lets call these sudoErrs for the rest of the ADR), it will handle the sudoErr in one of two ways.

### 1. Read the SudoErr - (*Default*)

The default behaviour is that the sudoErr will be stored in the module state for x number of blocks. The duration is dependent on a module parameter.

This sudoErr can be queried by the contract using the Stargate Querier.
```protobuf
service Query {
  // Errors queries all the errors for a given contract.
  rpc Errors(QueryErrorsRequest) returns (QueryErrorsResponse) {
    option (google.api.http).get = "/archway/cwerrors/v1/errors";
  }
}

message QueryErrorsRequest {
  // contract_address is the address of the contract whose errors to query for
  string contract_address = 1;
}

message QueryErrorsResponse {
  // errors defines all the contract errors which will be returned
  repeated SudoError errors = 1 [ (gogoproto.nullable) = false ];
}
```

### 2. Receive the SudoErr - (*Opt-in*)

The module exposes a subscription service where a contract can register a SudoErr Subscription for x number of blocks by paying y amount of tokens. The x and y values are dependent on module parameters. 


In case, a contract has the subscription, the sudoErr will be stored in the transient store. At the end of the block, all the sudoErrs are executed as `Sudo::Error` on the contract entrypoint. In case, this execution fails, the execution error will be stored in the state so as to not allow cascading error executions. This execution error will be wrapped in a new SudoError msg where the module_name would be `cwerror` and the error_code would be `ERR_CALLBACK_EXECUTION_FAILED` 

The max gas allowed for these sudoErr executions will be as small as possible while still be useful.  This execution is only meant for error handling and not for complex logic prosessing. 

The contract will receive a callback at the following entrypoint

```rust
#[cw_serde]
pub enum SudoMsg  {
    Error {
        module_name: String, // The name of the module which generated the error
        error_code: u32, // module specific error code
        contract_address: String, // the contract address which is associated with the error; the contract receiving the callback
        input_payload: String, // any relevant input payload which caused the error
        error_message: String, // the relevant error message
    }
}
```

A contract can subscribe to the error by registering with the following Msg

```protobuf
service Msg {
  // SubscribeToError defines an operation which will register a contract for a sudo callback on errors
  rpc SubscribeToError(MsgSubscribeToError) returns (MsgSubscribeToErrorResponse);
}

message MsgSubscribeToError {
  // sender is the address of who is registering the contarcts for callback on error
  string sender = 1;
  // contract is the address of the contract that will be called on error
  string contract_address = 2;
  // fee is the subscription fee for the feature (current no fee is charged for this feature)
  cosmos.base.v1beta1.Coin fee = 3 ;
}

message MsgSubscribeToErrorResponse {
  // subscription_valid_till is the block height till which the subscription is valid
  int64 subscription_valid_till = 1;
}
```

As mentioned above, the module params will be following:
```protobuf
message Params {
  // error_stored_time is the block height until which error is stored
  int64 error_stored_time = 1;
  // subsciption_fee is the fee required to subscribe to error callbacks
  cosmos.base.v1beta1.Coin subscription_fee = 2;
  // subscription_period is the period for which the subscription is valid
  int64 subscription_period = 3;
}
```

## Consequences

### Backwards Compatibility

Since the feature is being added as a new module, this should not cause any backwards compatibility issues.

### Positive

1.  Allows for future proofing against the growing codebase and having a dedicated error handling mechanism for sudo calls.
2.  Easier dapp development on top of Archway as only one error entrypoint needs to be implemented for all protocol invoked sudoErrs.
3.  Increases protocol revenue for the validators and delegators

### Negative

1.  Increases the exposure of custom wasm bindings of the protocol
2.  Introduces a layer of complexity to the protocol architecture

## Further Discussions

Future iterations of the module could include the following: 

1. Subscribe to errors from other contracts. e.g a dapp/multisig consists of many contracts working together. One contract can choose to receive errors about all its dependent contracts in one place.
2. Customize which module/error codes a contract wants a sudo call for, and which error codes its fine reading via stargate query.

## References

1. [RFC: x/cwerrors module](https://github.com/orgs/archway-network/discussions/35)
2. [AIP: x/cwerrors module](https://github.com/archway-network/archway/issues/544)
3. [SPEC: x/cwerrors module](../../x/cwerrors/spec/README.md)