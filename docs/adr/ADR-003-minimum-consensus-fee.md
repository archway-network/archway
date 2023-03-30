# ADR-003 Archway minimum consensus fee

## Status

Already Implemented.

## Abstract

This document proposes the introduction of a consensus fee, which defines the price for each unit of gas of a tx.

## Context

There are some transactions in which inflationary rewards + gas rebates can be higher than TX fees.
When this happens it means that a user (or validator) can include in a block a TX which costs, for example, 1 ARCH in fees but earns back 2 ARCH through gas rebates and inflationary rewards.
This is not a form of attack vector, but it is rather using some market inefficiency (similar to arbitraging) to make some profit.
This form of market inefficiency is brought back to balance by simply having more users posting TXs with higher fees to accrue for less and less rewards every time, until that profit becomes basically zero.

### The problem with producer extractable value inside archway
Validators which produce blocks can produce a block with only one TX, with gas wanted equal to the block gas and gas price equal to 0.
The TX can fail, computation is 0, fees paid for this are 0, and inflationary rewards are fully extracted.

## Proposal
The solution to make archway efficient and make it impossible for block producers to extract value is to introduce a minimum consesus fee.
This consensus fee creates a minimum gas price (ARCH required for one single unit of gas) inside the network.
NOTE: fees are sent to the `fee collector` which is then handled by distribution.

### Math
What we want to achieve: $${Inflationary Rewards + Gas Rebates \leq TX Fees}$$
Where: $${InflationaryRewards = \frac{GasConsumedByTX \times ArchInflationInBlock}{BlockGasLimit}}$$
Where: $${GasRebates = TxFees \times GasRebateRatio}$$
Where: $${TxFees=GasConsumedByTx \times GasPrice}$$

Let's better explain each component before moving to equality expansion and solution.

- InflationaryRewards: is the amount of inflationary rewards a dApp is given from the inflationary rewards the minting module generates which is allocated for dApps only.
- GasConsumedByTX: it's self explanatory, it defines the gas consumed by a TX when it runs (ex: MsgBankSend consumed 100_000 gas units)
- ArchInflationInBlock: defines the total archway inflationary rewards in a given block which are allocated to dApps.
- GasRebates defines the amount of fees which a user pays for a TX which are given to dApps. (ex: TX interacted with a contract, user paid 1ARCH in fees to run the TX then 0.5ARCH are given to the contract the TX interacted with)
- TxFees: self explanatory, they're the transaction fees set by the user.
- GasRebateRatio: defines the percentage of TX fees which need to be given to dApps. Ex: tx fees are 1 ARCH, GasRebateRatio=0.5, then 0.5ARCH (1 ARCH * 0.5) need to be given to dApps.
- GasPrice: defines the amount of coins that are paid for each unit of gas. (Ex: fees are 1 ARCH, gas consumed by tx is 1_000 GAS, then GasPrice is 1ARCH/1000 = 0.0001ARCH).

Now let's expand the first inequality:

$${\tiny{\frac{GasConsumedByTX \times ArchInflationInBlock}{BlockGasLimit} + GasConsumedByTx \times GasPrice \times GasRebateRatio \leq GasConsumedByTx \times GasPrice}}$$

Let's use shorthands for the variables at play:
- GasConsumedByTX = a
- ArchInflationInBlock = b
- BlockGasLimit = c
- GasPrice = d
- GasRebateRatio = e

$${\frac{a \times b}{c} + a \times d \times e \leq a \times d}$$

Now we solve the inequality for the gas price which is `d` in our inequality.
We want to solve the inequality for the `gas price` because we want to get a minimum fee that network can accept for each TX in a way for which the TX cannot accrue more rewards than what it is actually paying in fees (sybil attack vector).

$$
    \begin{equation}
        \begin{cases}
            a > 0\\
            b >  0\\
            c > 0\\
            d \geq 0\\
            0 \leq e \leq 1\\
            \frac{a \times b}{c} + a \times d \times e \leq a \times d\\
        \end{cases}       
    \end{equation}
$$

Which gets us to:

$$
    \begin{equation}
        \begin{cases}
            a > 0\\
            b > 0\\
            c > 0\\
            d \geq 0\\
            0 \leq e < 1\\
            d \geq - \frac{b}{c \times e - c}\\
        \end{cases}       
    \end{equation}
$$

The minimum consensus fee is represented by the following function: $${d \geq - \frac{b}{c \times e - c}}$$

## Implementation
Introduce an extra ante handler which once it knows the minimum fee for a TX computes if it meets the minimum consensus fee.
If it doesn't the TX fails for not meeting the minimum required fee.
