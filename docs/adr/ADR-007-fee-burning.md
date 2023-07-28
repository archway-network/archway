# ADR-007 – Introduction of fee burning

Date: 2023-07-27

## Status

Implemented

## Context

To better accommodate inflationary rewards increases within the Archway protocol, we are introducing transaction (TX)
fee burning. This strategy permits us to offer more inflationary rewards to contracts safely, while simultaneously
incinerating the TX fees which were previously distributed to validators and stakers.

## Decision

We plan to modify the Fee Deduction ante handler to burn the fees formerly allocated to validators and stakers. This 
change involves extending the behavior of the FeeDeduction ante handler. Currently, the FeeDeduction ante handler
dispatches the TX fees (which are not dedicated to contracts) to the fee collector module address. 
This address then becomes a rewards pool for staking rewards, managed by the Distribution module. Our new approach is to
burn these funds as soon as the FeeDeduction handler sends them to the fee collector.

## Consequences

### Positive
a) This change enables a safe increase of inflationary rewards for contracts, mitigating the risk of potential spam attacks and economic imbalances.
b) It establishes a transparent performance metric for Archway. The more fees burnt, the more value is generated for the protocol.

### Negative

a) Validators and stakers will experience a decrease in their rewards.
b) TX consume slightly more gas (12000gas more approximately), as they will need to pay for the gas used to burn the fees.
