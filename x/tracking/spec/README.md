<!--
order: 0
title: Tracking Overview
parent:
  title: "Tracking"
-->

# Tracking

## Abstract

The tracking module measures gas consumption per transaction.

> Only CosmWasm *Execute* and *Migrate* messages.

### Contract operations

Transactions could have multiple operations for one or more contracts (for example contract A calls contract B).
In order to persist this information, the [ContractOperationInfo](01_state.md#ContractOperationInfo) object is used.

> This object is pruned as soon as rewards are disbursed by the [x/rewards module](../../rewards/spec/README.md).

### Transaction info

In order to accurately measure gas consumption, each tracked transaction must have:
  - a unique ID increased sequentially;
  - height related to the block height of the transaction;
  - total gas represented by the sum of gas consumed in all contract operations:
  
      $$
        TotalGas  = GasSDK + GasVM
      $$

      where:
      * *GasSDK* - total gas used by the transaction outside of the VM;
      * *GasVM* - total gas used by contract within the CosmWasm VM;


### Tracking engine

The Archway protocol uses a CosmWasm modified `WasmerEngine` which utilizes a custom [Gas Processor](README.md#Gas processor) and is able to keep a record of a contract execution and subsequent executions.

### Gas processor

Intercepts smart contract operations and initializes the tracking of contract operation gas usage.

### Transaction tracking

Tx tracking happens as follows:

1. Tx is received by [ante handler](02_ante_handlers.md).
2. An empty [TxInfo](01_state.md#TxInfo) is created.
3. [Gas processor](README.md#Gas processor) creates a new [ContractOperationInfo](01_state.md#ContractOperationInfo).
4. [EndBlocker](03_end_block.md) finalizes tx tracking for the current block.

## Contents

1. **[State](01_state.md)**
2. **[Ante Handlers](02_ante_handlers.md)**
3. **[End-Block](03_end_block.md)**
4. **[Client](04_client.md)**
