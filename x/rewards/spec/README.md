<!--
order: 0
title: Rewards Overview
parent:
  title: "rewards"
-->

# Rewards

## Abstract

The module enables Cosmos SDK-based blockchain to calculate and distribute dApp rewards within the Archway protocol.
Module introduces a concept of *minimal consensus fee* to set the lower bound of a transaction fee. Module also introduces a concept of *contract flat fee* which a contract owner can set as the minimum fee that the contract expects to perform a transaction

### dApp rewards

dApp rewards are estimated based on a particular contract usage within one block using gas tracking data from the `x/tracking` module.
Basically, the more gas a contract uses (by its own or by other contracts calling it), the more rewards it gets.

To configure rewards specific parameters for a contract, the [ContractMetadata](01_state.md#ContractMetadata) object is used.

There are two types of dApp rewards a contract can receive.

#### Transaction fee rebate rewards

A portion of a transaction fee is used as a source of this reward. Rewards estimation formula:

$$
ContractFeeRewards = ( TxFees * TxFeeRebateRatio ) * \frac{ContractTxGasUsed}{TxGasUsed}
$$

where:

* *TxFees* - transaction fees paid by a user;
* *TxFeeRebateRatio* - `x/rewards` module parameter that defines the ratio to split fees between the **FeeCollector** and the **Rewards** module accounts (`[0..1)`);
* *ContractTxGasUsed* - total gas used by a contract within this transaction;
* *TxGasUsed* - total gas used by all contracts within this transaction;

> **FeeCollector**'s part of fees is used to reward validators and delegators as it is done in a "standard" Cosmos SDK-based chain. The same applies to inflationary rewards.

#### Inflationary rewards

A portion of minted by the `x/mint` module tokens is used as a source of these rewards. Rewards estimation formula:

$$
ContractInflationRewards = (MintedTokens * InflationRewardsRatio) * \frac{ContractTotalGasUsed}{BlockGasLimit}
$$

where:

* *MintedTokens* - amount of tokens minted per block by the `x/mint` module;
* *InflationRewardsRatio* - `x/rewards` module parameter that defines the ratio to split inflation between the **FeeCollector** and the **Rewards** module accounts (`[0..1)`);
* *ContractTotalGasUsed* - total gas used by a contract within this block;
* *BlockGasLimit* - maximum gas limit per block (consensus parameter);

#### Transaction fees

$$
MinimumTxFee = (MinConsensusFee * TxGasLimit) + \sum_{msg=1, type_{msg} = MsgExecuteContract}^{len(msgs)} flatfee(ContractAddress_{msg})
$$

where:

* $MinimumTxFee$ - minimum fees expected to be paid for the given transaction;
* $MinConsensusFee$ - price for one gas unit;
* $TxGasLimit$ - transaction gas limit provided by a user;
* $ContractAddress_{msg}$ - contract address of the msg which needs to be executed;
* $flatfee(x)$ - function which fetches the flat fee for the given input;

##### Minimum consensus fee

The *minimum consensus fee* is a price for one gas unit. That value limits the minimum fee paid by a user in respect to the provided transaction gas limit:

The *minimum consensus fee* value is updated each block using the formula:

$$\displaylines{
MinConsensusFee = -\frac{InflationBlockRewards}{BlockGasLimit * TxFeeRebateRatio - BlockGasLimit} \\
InflationBlockRewards = MintedTokens * InflationRewardsRatio
}$$

##### Contract flat Fee

The *contract flat fee* is a fee set by the contract owner. Any user executing a msg on that contract needs to pay this amount as part of their transaction fee. When a transaction has multiple messages which call different contracts, the flat fees for all the contracts need to be paid. Contract owners  can choose any native token as their contract flat fees, it does not have to be the default token of the chain.

> If the provided transaction fee is less, then MinimumTxFee, transaction is rejected.
> User can estimate a transaction fee using the `x/rewards/EstimateTxFees` query.

## Contents

1. **[State](01_state.md)**
2. **[Messages](02_messages.md)**
3. **[Ante Handlers](03_ante_handlers.md)**
4. **[End-Block](04_end_block.md)**
5. **[Events](05_events.md)**
6. **[Parameters](06_params.md)**
7. **[Client](07_client.md)**
8. **[WASM bindings](08_wasm_bindings.md)**
