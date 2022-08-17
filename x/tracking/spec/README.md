<!--
order: 0
title: Tracking Overview
parent:
  title: "Tracking"
-->

# Tracking

## Abstract
The tarcking module measures gas consumption per transaction.

> Only CosmWasm Execute & Migrate sdk.Msgs

### Contract Operation Info
Transactions could have multiple operations for one or more contracts (for example contract A calls contract B).
In order to persist this information `ContractOperationInfo`.
> This object is pruned as soon as rewards are disbursed by the [rewards module](../../rewards/spec/README.md)

### Transaction Info
In order to accurately measure gas consumption each tracked trasnaction must have:
  - A unique ID increased sequentially
  - height related to the block height of the transaction
  - total gas represneted by the sum of gas consumed in all contract operations
      $$
        TotalGas  = GasSDK + GasVM
      $$

      where:
      * *GasSDK* - total gas used by the transaction outside of the VM
      * *GasVM* - total gas used by contract within the CosmWasm VM


### TrackingEngine
A cosmwasm modified WasmerEngine which utilizes a custom [Gas Processor](README.md#GasProcessor), and is able to keep a record of an execution and subsequent executions.

### Gas Processor
Intercepts smart contract operations and initializes tracking of contract operations.

### Transaction Tracking
Tx tracking happens as follows:

1. Tx is received by [ante handler](03_ante_handlers.md)
2. An empty [TxInfo](01_state.md#TxInfo) is created.
3. [Gas Processor](README.md#GasProcessor) creates new [ContractOperationInfo](01_state.md#ContractOperationInfo).
4. [EndBlocker](04_end_block.md) performs finalizes tx tracking for the block.


## Contents

1. **[State](01_state.md)**
2. [Ante Handlers](02_ante_handlers.md)
3. **[Begin-Block](03_end_block.md)**
4. **[Client](04_client.md)**


