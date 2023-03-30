# ADR 004 - Contract premiums

## Status

Already implemented.

## Abstract

Contract premiums allow smart contract developers to define a custom flat fee for interacting with their smart contract.

## Context

Contract developers can use contract premiums to define a custom fee, that is applied after computational fees.

Contract premiums can be used to cover hidden costs of a smart contract, for example a NFT marketplace which delivers goods
can use contract premiums to cover delivery costs.

The reason for which contract premiums are useful instead of using `x/wasm` `funds` is because of fee predictability.
In fact `Contract Premiums` define a standardised way to define contract custom fees and can be used by wallets to predict fees.

### Proposal

We add a new `sdk.Msg` to `x/rewards` called `MsgSetFlatFee` which allows the contract `metadata` owner to define a custom
flat fee.

We then extend our `FeeDeduction` `AnteHandler` to fetch the `FlatFee` of a contract, if the `FlatFee` exists then the 
`AnteHandler` ensures the `tx.Fees` are enough to also cover the `FlatFee`, making tx costs explicit for the end-user too.
`FlatFees` are then sent directly to the contract's `metadata.RewardAddress`.

#### Limitations

The `FlatFee` is imposed only on the first contract call, which means they're imposed when there are interactions between 
externally owned accounts and contracts. They're not imposed in contract to contract interactions, this is not to hinder 
fee predictability.

This means that if a contract is called and has a flat fee set, then the contract **MUST** check itself if the sender is 
an externally owned account or a contract and apply the flat fee accordingly.

The protocol defines efficient wasm bindings for querying the flat fees of a contract, such that this information can be used
by contracts to force flat fees even when the caller is a contract.
