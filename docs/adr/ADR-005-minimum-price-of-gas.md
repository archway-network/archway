# ADR-005: Introducing Minimum Price of Gas for Transaction Fees

## Status

Implemented


## Abstract

This ADR proposes the introduction of a minimum price of gas concept to define the fees in the network. Transaction fees
will be determined by the minimum price of gas multiplied by the transaction gas limit. This will ensure that developer
rewards remain meaningful over the long term and prevent validators from accepting transactions with extremely low fees.

## Context

Currently, we have the concept of Minimum Consensus Fee, which serves as a protection against DoS attacks by ensuring
that the protocol cannot payout more in rewards than what it is getting as fees. However, this is not intended to
function as a fee market mechanism.

The need for a better mechanism to manage the fee market has arisen in order to ensure that developer rewards are material
and to allow for adjustments to minimum protocol fees as the price of Archway changes.

## Proposal

We propose the implementation of a Minimum Price of Gas concept, which will be used to define the fees in the network.
This will be achieved by setting transaction fees equal to the minimum price of gas multiplied by the transaction gas
limit. Both the minimum price of gas and the minimum consensus fee will be taken into consideration, with the highest
value being used.

### Key Features:

1. Ensures developer rewards remain meaningful in the long term.
2. Prevents validators from accepting transactions with extremely low fees.
3. Allows for adjustments to minimum protocol fees based on the price of Archway.
4. Considers both minimum price of gas and minimum consensus fee, using the highest value.
5. The parameter can be changed through governance.
6. In the future, the parameter will be dynamically updated by the protocol based on different factors, such as block fill percentage.

### Implementation Steps:

1. Add a MinPriceOfGas parameter to the x/rewards module parameters.
2. Define the initial value for minimum price of gas.
3. Modify the x/rewards fee management ante handler to account for the minimum price of gas.
4. Update the x/rewards EstimateTxFees endpoint to account for the minimum price of gas.

Once implemented, this proposal will allow for a more effective management of the fee market, ensuring that developer
rewards remain material and enabling adjustments to minimum protocol fees based on the price of Archway. Furthermore, 
it lays the groundwork for future dynamic updates of the parameter based on factors like block fill percentage.
