<!--
order: 0
title: Rewards Overview
parent:
  title: "rewards"
-->

# `Rewards`

## Overview

The module enables Cosmos-SDK based blockchain to calculate and distribute dApp rewards within the Archway protocol.
Module also introduces a concept of *minimal consensus fee* to set the lower bound of a transaction fee.

### dApp rewards

Contract rewards are estimated based on a particular contract usage within one block using gas tracking data from the `x/tracking` module.
Basically, the more gas a contract use (by its own or by other contracts calling it), the more rewards it gets.

To configure rewards specific parameters for a contract, the [ContractMetadata](01_state.md#ContractMetadata) object is used.

There are two types of dApp rewards a contract could receive.

#### Transaction fee rebate rewards

A portion of a transaction fee is used as a source of this reward. Rewards estimation formula:

$$
ContractFeeRewards = ( TxFees * TxFeeRebateRatio ) * \frac{ContractTxGasUsed}{TxGasUsed}
$$

where:

* *TxFees* - transaction fees payed by a user;
* *TxFeeRebateRatio* - `x/rewards` module parameter that defines the ratio to split fees between the **FeeCollector** and the **Rewards** module accounts (`[0..1)`);
* *ContractTxGasUsed* - total gas used by a contract within this transaction;
* *TxGasUsed* - total gas used by all contracts within this transaction;

> **FeeCollector**'s part of fees is used to reward validators and delegators as it is done in a "standard" Cosmos-SDK based chain. The same applies to the inflationary rewards.

#### Inflationary rewareds

A portion of minted by the `x/mint` module tokens is used as a source of this rewards. Rewards estimation formula:

$$
ContractInflationRewards = (MintedTokens * InflationRewardsRatio) * \frac{ContractTotalGasUsed}{BlockGasLimit}
$$

where:

* *MintedTokens* - amount of tokens minted per block by the `x/mint` module;
* *InflationRewardsRatio* - `x/rewards` module parameter that defines the ratio to split inflation between the **FeeCollector** and the **Rewards** module accounts (`[0..1)`);
* *ContractTotalGasUsed* - total gas used by a contract within this block;
* *BlockGasLimit* - maximum gas limit per block (consensus parameter);

#### Minimum consensus fee

The *minimum consensus fee* is a price for one gas unit. That value limits the minimum fee payed by a user in respect to the provided transaction gas limit:

$$
MinimumTxFee = MinConsensusFee * TxGasLimit
$$

The *minimum consensus fee* value is updated each block using the formula:
$$
MinConsensusFee = -\frac{InflationBlockRewards}{BlockGasLimit * TxFeeRebateRatio - BlockGasLimit} \\
InflationBlockRewards = MintedTokens * InflationRewardsRatio
$$

> If the provided transaction fee is less then MinConsensusFee x TxGasLimit, transaction is rejected.
> User could estimate a transaction fee using the `x/rewards` query.

## Contents

1. **[State](01_state.md)**
2. **[Messages](02_messages.md)**
3. [Ante Handlers](03_ante_handlers.md)
4. **[Begin-Block](04_begin_block.md)**
5. **[Events](05_events.md)**
6. **[Parameters](06_params.md)**
7. **[Client](07_client.md)**
