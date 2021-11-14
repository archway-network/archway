#Gastracker

## Abstract
This document specifies the gas tracker wrapper for the archway network.

This module is responsible for intercepting CosmWasm VM messages, measuring gas consumption per call.

## Contents
1. **[State](./01_state.md)**
2. **[Keeper](./02_keepers.md)** 
  - **[BaseKeeper](./02_keepers.md#Keeper)** 
3. **[Messages](03_messages.md)**
4. **[Events](04_events.md)**

## Concepts
### Gas Tracking
Gas tracking is the ability to observe transaction interactions between dApps and the blockchain, by gas tracking the protocol can determine how a dApp contributes to the network and provide rewards to the apps that help the attain the most network gain.

### Contract Metadata
A subset of data that allows categorization and functionality, for archway puposese the only metadata stored are: 
-   `reward_address`
-   `gas_rebate_to_user`
-   `collect_premium`
-   `premium_percentage_charged`

### Gas Rebates
A small amount distributed to both validators and dApps from each transaction executed on the chain.

### Contract Premium
Contract premiums allow to charge a higher amounts of gas that will be distribured to the contract `reward_address`.
