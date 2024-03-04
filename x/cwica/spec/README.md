# CWICA

This module enables Cosmwasm based smart contracts to register ICA accounts and submit transactions to be executed on counterparty chains.

## Concepts

Interchain Accounts is the Cosmos SDK implementation of the ICS-27 protocol, which enables cross-chain account management built upon IBC. Unlike regular accounts, interchain acocunts are contolled programatically by the smart contracts on archway via IBC packets.

The module has been designed such that a single smart contract can have one account per ibc connection.

You can find more docs about ICA [here](https://ibc.cosmos.network/main/apps/interchain-accounts/overview) and [here](https://github.com/cosmos/ibc/blob/main/spec/app/ics-027-interchain-accounts/README.md)

## How to use in CW contract

The ICA functionality is broadly abstracted away to make development easier for dapp developers. 

### Operations

#### Register Interchain Account

The contract can register an interchain account as is shown in the following snippet.

```rust
let regsiter_msg = MsgRegisterInterchainAccount {
    from_address: env.contract.address.to_string(), // the smart contract address
    connection_id: connection_id, // the IBC connection id which will be used to create the interchain accountzd
};

let register_stargate_msg = CosmosMsg::Stargate { 
    type_url: "/archway.cwica.v1.MsgRegisterInterchainAccount".to_string(),
    value: Binary::from(prost::Message::encode_to_vec(&regsiter_msg)),
};

Ok(Response::new().add_message(register_stargate_msg))
```

Once the interchain account is generated on the counterparty chain, the contract will receive a callback at the following Sudo endpoint.  The address of the account on the counterparty chain is available in `counterparty_version`.

```rust
#[cw_serde]
pub enum SudoMsg  {
    Ica {
        account_registered: Option<AccountRegistered>, // This endpoint is hit on the creation of the interchain account. It will also include the address on the counterparty chain
    },
}
```

#### Submit Txs

Once an interchain account is created, the contract can submit txs to be executed on the counterparty chain as is shown in the following snippet.

```rust
let vote_msg = MsgVote { // this example votes on behalf of the counterparty account
    proposal_id: proposal_id, // the governance proposal id
    voter: ica_address, // the address on the conterparty chain
    option: option, // the vote option
};

let vote_msg_stargate_msg = prost_types::Any { // proto encoding the MsgVote
    type_url: "/cosmos.gov.v1.MsgVote".to_string(),
    value: vote_msg.encode_to_vec(),
};

let submittx_msg = MsgSubmitTx {
    from_address: env.contract.address.to_string(), // the smart contract address
    connection_id: connection_id, // the ibc connection used when creating the ica
    msgs: vec![vote_msg_stargate_msg], // all the msgs to execute on the counterparty chain
    memo: "sent from contract".to_string(), // tx memo
    timeout: 200, // timeout in seconds
};
        
let submittx_stargate_msg = CosmosMsg::Stargate {
    type_url: "/archway.cwica.v1.MsgSubmitTx".to_string(),
    value: Binary::from(prost::Message::encode_to_vec(&submittx_msg)),
};
        
Ok(Response::new().add_message(submittx_stargate_msg))
```

Once the txs have been submitted, the contract will receive a callback at the Sudo entrypoints. [More info](../../../proto/archway/cwica/v1/sudo.proto)


> **NOTE** 
> 
> Please note that packet timeouts cause the ibc channel to be closed. The channel can be reopened again by registering the ica account again.

## Contents

1. [State](./01_state.md)
2. [Messages](./02_messages.md)
3. [Client](./03_client.md)
4. [Wasm Bindings](./04_wasm_bindings.md)