<!--
order: 0
title: Mint Overview
parent:
  title: "mint"
-->

# Mint

## Abstract

The module enables Cosmos SDK-based blockchain to calculate inflation for the current block, mint it and then distribute it to configured recipients.

### The Minting Mechanism
The minting mechanism was designed to:

* allow for a flexible inflation rate determined by market demand targeting a particular bonded-stake range
* effect a balance between market liquidity and staked supply
* based on a time based inflation scheme instead of a block based inflation scheme

In order to best determine the appropriate market rate for inflation rewards, a moving change rate is used. The moving change rate mechanism ensures that if the % bonded is either over or under the goal %-bonded, the inflation rate will adjust to further incentivize or disincentivize being bonded, respectively. Setting the goal %-bonded at less than 100% encourages the network to maintain some non-staked tokens which should help provide some liquidity.

It can be broken down in the following way:

* If the inflation rate is below the min %-bonded the inflation rate will increase until its between the accepted range
* If the goal % bonded is maintained, then the inflation rate will stay constant
* If the inflation rate is above the max %-bonded the inflation rate will decrease until its between the accepted range

The amount of tokens minted each block is dependent on the amount of time which has passed since the last time inflation was minted. 
To ensure that events like chain upgrade, chain halts do not cause large amounts of new tokens minted due to large time difference, there is a max block duration time param which caps the max token which can be minted.

## Contents

1. **[State](01_state.md)**
2. **[Begin-Block](02_begin_block.md)**
3. **[Events](03_events.md)**
4. **[Parameters](04_params.md)**
5. **[Client](05_client.md)**
