# ADR 004 - Contract Premiums

## Status

Already implemented.

## Abstract

Contract premiums allow smart contract developers to define a custom flat fee for interacting with their smart contract.

## Context

Contract developers can use contract premiums to define a custom fee, that is applied after computational fees.

Contract premiums can be used to cover hidden costs of a smart contract, for example a NFT marketplace which delivers goods
can use contract premiums to cover delivery costs.

The reasons for using contract premiums over using [`x/wasm funds`](https://book.cosmwasm.com/basics/funds.html) are:
1. Fee predictability: Contract Premiums define a standardized way to define contract custom fees and can be used by wallets to predict fees
2. Rewards on Msg Fail: When using Contract Premiums rewards will be distributed even when the contract msg execution fails. Using the x/wasm funds way would not reward the developer if the msg execution failed due to bad input by the user.
3. Rewards withdrawal: Contract Premiums sends all the rewards to the configured rewards address. Using the x/wasm funds option would send all the funds to the smart contract unless custom transfer logic is implemented.
4. Easier regulatory compliance: Using contract premiums, developer receives the rewards only when they explicitly request to withdraw (similar to how staking rewards works). Using x/wasm funds to receive the funds, which happens immediately, might complicate the tax situation based on the jurisdiction. 
5. One configuration to rule them all: Once set, Contract Premiums are applied to all Msg Executions exposed by the contract, as opposed having to be configured for every msg.

### Proposal

We add a new `sdk.Msg` to `x/rewards` called `MsgSetFlatFee` which allows the contract `metadata` owner to define a custom
flat fee.

We then extend our `FeeDeduction` `AnteHandler` to fetch the `FlatFee` of a contract, if the `FlatFee` exists then the 
`AnteHandler` ensures the `tx.Fees` are enough to also cover the `FlatFee`, making tx costs explicit for the end-user too.
`FlatFees` are then sent directly to the contract's `metadata.RewardAddress`.

### Limitations

#### FlatFee is imposed only on interactions between EOA and contract, not between contract and contract.

The `FlatFee` is imposed only on the first contract call, which means they're imposed when there are interactions between 
externally owned accounts and contracts. They're not imposed in contract to contract interactions, this is not to hinder 
fee predictability. In fact, considering contract interactions can be conditional and the condition can change on a block by 
block basis, the final fee would change based on these conditions, making the fees unpredictable.


Example when the call `ContractB` condition is `true`:

```mermaid
sequenceDiagram
    User->>ContractA: FlatFee Applied: 1ARCH
    ContractA->>ContractB: FlatFee applied: 2ARCH
    ContractB->>FinalFee: 3ARCH
```

Example when the call `ContractB` condition is `false`:
```mermaid
sequenceDiagram
    User->>ContractA: FlatFee Applied: 1ARCH
    ContractA->>FinalFee: 1ARCH
```


This means that if a contract is called and has a flat fee set, then the contract **MUST** check itself if the sender is 
an externally owned account or a contract and apply the flat fee accordingly.

The protocol defines efficient wasm bindings for querying the flat fees of a contract, such that this information can be used
by contracts to force flat fees even when the caller is a contract.

#### Reverts cause the FlatFee to be lost

On contract call failures the TX is reverted and the flat fee would be lost too. This is a limitation of the `cosmos-sdk`
that does not allow us to give the user the `FlatFee` back in case of TX failure as the SDK does not implement post tx execution 
handlers.


### User Experience – A note on wallet and frontend integration

Contract premiums and minimum consensus fees only affect transactions that involve WASM contract execution. They don't 
change other processes like staking, governance, transfers, and so on.

Still, fees for contract interactions need to be changed, and we'll explain how to do that below.

First, wallets don't need to be changed to work with Archway Network's special fees, since normal operations aren't affected 
by this fee management system. So, only contract interactions need to be handled.

A contract always has a user interface (UI). UIs usually work as a go-between for a wallet (like Keplr) and the contract.
The front-end, or what users see and interact with, is the part that needs to be changed to handle contract premiums correctly. 
This is fair because the contract developer, who sets the contract premium, is also the one who created the contract.

#### Correct fee setting flow

Once the UI knows which is the message that needs to be sent to the contract, it needs to set the fees for the TX,
in order to correctly set fees it needs to:
1. Simulate the TX, using the standard Simulate TX endpoint of `cosmos-sdk`, this returns the estimated `gas_limit` for the TX.
2. Send a query to the archway [EstimateTxFees](../../proto/archway/rewards/v1/query.proto?L32) 
gRPC query method, and feed it the `gas_limit` returned in step `1.` and the contract being interacted with. 
3. Set the fee in the wallet TX.
