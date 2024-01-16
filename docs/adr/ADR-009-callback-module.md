# ADR-009 – Callback module

Date: 2023-01-11

## Status

Implemented

## Abstract

We propose a x/cw-callback module which enables CosmWasm smart contracts to receive callbacks at the end of a block.
This is useful for scheduling actions to happen at an expected time by reserving block execution in advance.

## Context

Currently on Archway, the smart contracts can only be executed when called on by a user or by other contracts. The decentralized nature of ecosystem should allow enforcement of predefined protocols without the need for intermediaries.

However, for some applications, dapps still rely on trusted third parties to instigate the executions or state changes. Examples include, epochs for defi primitives, price balancing arbitrage bots, vesting distribution of funds, auto compounding of rewards. In case of Archway, with its unique rewards mechanism, the withdrawal of contract rewards would need to happen on a recurring basis.

The aim of this AIP is to solve the reliance on the trusted third parties for these types of scripting reasons. The protocol should enable a minimal scheduler service which can be used to handle such operations in a permissionless and trust minimized way.

## Alternatives

There are a few existing solutions within the Cosmos ecosystem which are similar to what we aim to introduce. They are elaborated below.

### 1. Stargaze's x/cron & Juno's x/clock modules

Stargaze v12 has the [x/cron](https://github.com/public-awesome/stargaze/tree/v12.0.0/x/cron) module and Juno v17 implements the [x/clock](https://github.com/CosmosContracts/juno/tree/v17.0.0/x/clock) module. Both operate by pinging configured contracts every block. They have a governance process which allows a dapp developer to add their contract address to the EndBlocker hooks. Therefore, at the end of every block, the chain calls the contract and the contract can handle this trigger however it wishes. This happens in perpetuity unless a governance process removes the contract from receiving the callbacks.

We don't believe this works for Archway for the following reasons:
1. Making the feature governance gated gives an unfair advantage to incumbents and makes it harder for new dapps (without a strong community or userbase yet) to be able get their dapps accepted into the callback pool.
2. Due to Archway being a permissionless contract chain, we have no control over the contracts which are uploaded. As such we can't trust any contract address to be given perpetual callback access in every block not knowing what it does or if it migrates to a malicious contract in the future.
3. These callbacks are called by the protocol and as such the contract gets free execution[^1] for this msg execution. As such, this would be unfair to the rest of the contracts on the chain which are making their users pay for their actions while their competitors are getting the benefit of free execution[^1].

Interesting takeaways:
1. Juno's x/clock module has a module parameter `ContractGasLimit` which is a governance-controlled parameter which controls the maximum gas which can be consumed by the contracts in a given block. This prevents contracts from abusing the free execution[^2]. 

### 2. Neutron's x/cron 

Neutron's [x/cron](https://github.com/neutron-org/neutron/tree/v1.0.4/x/cron) module is similar to the above solutions in the sense that it is governance gated. 
However, this module allows custom cron schedule and specifies a list of CosmWasm msgs which will be called with the execution. 

Along with the reasons specified for the previous modules, we don't believe this is ideal for Archway either, due to
1. This still calls the schedule in perpetuity. The possibility of migrating the contracts in the msgs to a malicious one still persists.
2. Storing the execution msgs is also something the module does for free. Though this might not be significant, the storage costs of the schedule should not be borne by the chain.
3. While Neutron's x/cron module is a lot more powerful, it does limit itself on how many can actually leverage the feature as there is a hard cap on how many users can get this access.

Interesting takeaways:

1. Neutron's x/cron module using a user provided cron schedule to perform the callbacks ensures that the contracts are receiving the callback only when they want and that computation is not being spent when the contracts might not need it. 
2. The module has a parameter `limit` which is a governance-controlled parameter which controls the number of schedules which can be handled by the module.

## Decision

We will implement a new module called x/cw-callback which solves the specified problem and enables automated, time triggered actions. With this, a dapp will be able to ask the chain to give it a callback at a specific height so its can perform any desired operations. The field of job_id is added to allow the contract to have context on the callback.

```proto
message Callback {
  string contract_address = 1;
  uint64 job_id = 2;
  uint64 callback_height = 3;
}
```

The module relies on the ABCI `EndBlocker` to execute these callbacks. Each callback will be called with a custom context and any errors thrown by it during the execution, will be collected and thrown as events. This is due to the fact that a wallet is not instigating this execution, to return the error using the execution context. 

As the execution of the callbacks is done by the chain, it is the validators who are paying for the execution of the callback. Since they will not be receiving the transaction fees for the executions, explicit value of transaction fees will be paid by the contract when requesting the callback. Because the module will not be aware of how much gas a callback will consume, the contract is expected to overpay the transaction fees in advance and the contract will receive any leftover gas fees post callback execution. If the callback is cancelled, this amount is returned completely.

The total amount to be paid for a callback is calculated as:

$callbackFee = transactionExecutionFee + totalReservationFee$

$transactionExecutionFee = callbackGasLimit_{params} \times estimateFees(1)$

$totalReservationFee = (resevationFee_{blockReservation} + resevationFee_{futureReservation})$

$resevationFee_{blockReservation}= (maxBlockReservationLimit_{params}  - count(callbacks_{currentHeight})) \times blockReservationFeeMultiplier_{params}$	

$resevationFee_{futureReservation}= (blockHeight_{callback} - blockHeight_{current}) \times futureReservationFeeMultiplier_{params}$


The module will have the following module params:
```proto
message Params {
    uint64 callback_gas_limit = 1;
    uint64 max_block_reservation_limit = 2;
    uint64 max_future_reservation_limit = 3;
    string block_reservation_fee_multiplier = 4 [(gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec"];
    string future_reservation_fee_multiplier = 5 [(gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec"];
}
```

1. callback_gas_limit 

  This is the maximum gas that can be consumed by a callback. If the gas consumed exceeds this value, the callback will fail automatically. This ensures, that the contracts aren't abusing the callback feature to perform high gas computations. This value is expected to be large enough to allow meaningful executions but small enough to prevent abuse.
  
2. max_block_reservation_limit 

  This is the maximum number of callbacks which can be registered in a given block. If this value is 5, maximum of 5 callbacks can be registered for a given block height. This prevents too many callbacks from slowing down the performance of the chain.
  
3. max_future_reservation_limit 

  This is the maximum number of blocks in the future that a contract can request a callback in. If this value is 100, the contracts can only request callbacks in the next 100 blocks. The callback feature provides guaranteed block execution; therefore, this limit is necessary to ensure the callbacks aren't reserved without meaningful intention. 

4. block_reservation_fee_multiplier 

  This is a value which calculates a part of the reservation fees which will need to be paid when requesting the callback. As the number of callbacks for a given block approach `max_block_reservation_limit`, this multiplier is applied as it will disincentivize too many callbacks from being registered in a single block.
  
5. future_reservation_fee_multiplier 

  This is a value which calculates a part of the reservation fees which will need to be paid while requesting the callback. The difference between callback request height and current block height will be multiplied by this value to calculate the reservation fees for the callback. This reduces the incentives to request callbacks far in the future, as a way to hedge against any raising gas prices.

The module will expose the following msg services:
```proto
service Msg {
  // Updates the module pararmeters
  rpc UpdateParams(MsgUpdateParams) returns (MsgUpdateParamsResponse);
  // Requests a new callback
  rpc RequestCallback(MsgRequestCallback) returns (MsgRequestCallbackResponse);
  // Cancels an existing callback
  rpc CancelCallback(MsgCancelCallback) returns (MsgCancelCallbackResponse);
}

message MsgUpdateParams {
  option (cosmos.msg.v1.signer) = "authority";
  string authority = 1;// authority is the address that controls the module (defaults to x/gov unless overwritten).
  Params params = 2 [(gogoproto.nullable) = false, (gogoproto.jsontag) = "params,omitempty"];
}

message MsgUpdateParamsResponse {}

message MsgRequestCallback {
  option (cosmos.msg.v1.signer) = "sender";
  string sender = 1;
  uint64 job_id = 2;
  uint64 callback_height = 3;
  cosmos.base.v1beta1.Coins fees = 4;
}

message MsgRequestCallbackResponse {}

message MsgCancelCallback{
  option (cosmos.msg.v1.signer) = "sender";
  string sender = 1;
  uint64 job_id = 2;
  uint64 callback_height = 3;
}

message MsgCancelCallbackResponse {} 
```

The module will expose the following queries:
```proto
service Query {
  // Calculate how much callback fees a contract needs to pay to register the callback
  rpc EstimateCallbackFees(QueryEstimateCallbackFeesRequest)
      returns (QueryEstimateCallbackFeesResponse) { }
  // Returns all the calbacks registered at a given height
  rpc Callbacks(QueryCallbacksRequest)
      returns (QueryCallbacksResponse) { }
}

message QueryEstimateCallbackFeesRequest{
  uint64 block_height = 1;
}

message QueryEstimateCallbackFeesResponse{
  cosmos.base.v1beta1.Coins totalFees = 1;
  cosmos.base.v1beta1.Coins transactionFees = 2;
  cosmos.base.v1beta1.Coins blockReservationFees = 3;
  cosmos.base.v1beta1.Coins futureReservationFees = 4;
}

message QueryCallbacksRequest{
  uint64 block_height = 1;
}

message QueryCallbacksResponse{
  repeated Callback callbacks = 1;
}
```

The module will expose the following wasm bindings:
```rust
#[cw_serde]
pub enum SudoMsg {    
    Callback { job_id: u64 },
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn sudo(_deps: DepsMut, _env: Env, msg: SudoMsg) -> Result<Response, ContractError> {
    match msg {        
        SudoMsg::Callback { job_id } => {
            Ok(Response::new())
        }
    }
}
```

## Consequences

### Backwards Compatibility

Since the feature is being added as a new module, this should not cause any backwards compatibility issues.

### Positive

1.  Allows contracts to be triggered without dependencies on third party bot solutions which could lead to unique protocols being built on top of Archway
2.  Sets groundwork for more comprehensive cron like solutions to be built as a service on the contract layer by other contracts
3.  Increases protocol revenue for the validators and delegators

### Negative

1.  Increases the exposure of custom wasm bindings of the protocol
2.  In case, the fees value increases significantly in the future, the validators encounter a loss as they were paid lesser fees when the callback was reserved

## Further Discussions

Future iterations of the module could include the following: 

1. Allow a contract to pay extra "incentives" to prioritize their callback at the given height.
2. Store the contract errors in the state for `n` blocks such that the developers will not need to set up event monitoring to access what the error was.
3. Algorithmically set the multipliers which are used in reservation fee calculation instead of relying on governance to set the params in a market efficient way.
4. Allow retry of the callback in the next block if the callback failed 


## References

[^1]: There is no free execution on the blockchain. By free execution, we mean that the validators are paying for the execution.