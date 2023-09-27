# ADR-008 – Refined Withdrawal User Experience

Date: 2023-09-27

## Status

Proposed

## Context

Developer feedback suggests that the process of withdrawing gas and inflationary rewards on Archway is currently cumbersome
and expensive. The existing process necessitates a contract to process a series of reward records, leading to substantial
gas charges. Implementing this consumption process at the contract level is non-trivial.

While the present reward withdrawal mechanism offers precision and supports the development of intricate distribution 
contracts, many use-cases do not demand such detailed control. Developers often prefer a simpler approach where rewards
are directly sent to a singular address.

## Decision

A new attribute named `distribute_to_wallet` will be introduced in the `ContractMetadata`. When this attribute is
activated (set to true), instead of generating a `RewardRecord` for the contract upon accruing gas or inflationary
rewards, the rewards will be directly dispatched to the `ContractMetadata.withdraw_address`.

- If `distribute_to_wallet` is set to false, the `RewardRecord` will be generated, necessitating manual withdrawal.
- If `distribute_to_wallet` is set to true, the `RewardRecord` will not be generated, and rewards will be directly sent
to the `ContractMetadata.withdraw_address`.

A new message named `MsgDistributeToWallet` will be incorporated into the `x/rewards` module, enabling the
`ContractMetadata` owner to toggle the `distribute_to_wallet` flag on or off.

## Consequences

### Positive

1. Eliminates the need for state migrations which, in the case of rewards, could be time-consuming due to extensive reward
state data.
2. Facilitates the creation of complex reward distribution contracts.
3. Simplifies scenarios where a developer merely desires the funds to be sent to a specific address.
4. Allows the owner to switch between automatic and manual distribution whenever necessary.

### Neutral

1. Existing `RewardRecords` will still require manual consumption.

### Negative

No known negative consequences at this time.
