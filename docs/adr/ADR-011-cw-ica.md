---

# ADR-011: Introduction of CW ICA Module

Date: 2024-03-19

## Status

Accepted | Implemented.

## Context

The introduction of the CWICA module allows the smart contracts on Archway to create accounts on other chains on behalf of users/other contracts as long as they are IBC connected and ICA enabled. Unlike native account which are controlled by private keys, these accounts are controlled by smart contracts.

## Decision

The module exposes two endpoints for the account management.

### 1. Create an account

An account can be created on the chain by executing the MsgRegisterInterchainAccount from the contract. This is done by Stargate msgs on Cosmwasm.

```protobuf
message MsgRegisterInterchainAccount {
  // contract_address is the address of the contract who wants to register an ica
  // account on the counterparty chain
  string contract_address = 1;
  // connection_id is the ibc connection id between the two chains
  string connection_id = 2;
}
```

When an account has been created, the contract will receive a callback with the address of the account on the counterparty chain.

The callback is received by the contract at the sudo entrypoint

```rust
pub enum SudoMsg  {
    Ica {
        account_registered: AccountRegistered,
    }
}

pub struct AccountRegistered {
    pub counterparty_address: String,
}
```

### 2. Execute txs with that account

Once an account has been created, the contract can submit txs on behalf of the interchain account. This too is done by Stargate msgs on Cosmwasm.

```protobuf
message MsgSendTx {
  // contract_address is the address of the contract who wants to submit a transaction to
  // the counterparty chain
  string contract_address = 1;
  // connection_id is the ibc connection id between the two chains
  string connection_id = 2;
  // msgs are the proto encoded messages to be submitted to the counterparty chain
  repeated google.protobuf.Any msgs = 3;
  // memo is the memo to be included in the packet
  string memo = 4;
  // timeout in seconds after which the packet times out
  uint64 timeout = 5;
}
```

On successful execution of the ICA tx, a contract will receive a callback with the response from the counterparty chain(if any).

The callback is received by the contract at the sudo entrypoint
```rust
pub enum SudoMsg  {
    Ica {
        tx_executed: ICAResponse,
    }
}

pub struct ICAResponse {
    pub packet: ibc.Packet,
    pub data: Binary,
}
```

The execution can fail due to two reasons,
1. Execution Failure: This could be due to the logical error or the protobuf msg encoding error. In this case, the contract will have an error with error code [`2`](../../proto/archway/cwica/v1/errors.proto)
2. Packet timeout: This could be due to the SendTx packet not being picked up before timeout. In this case, the contract will have an error with error code [`1`](../../proto/archway/cwica/v1/errors.proto). 

> **NOTE**
>
> In case of packet timeouts, the channel is closed (as ICA uses ORDERED channels). The channel can be reopened by registering the ICA account again.

In both the cases, the errors can be retrived via the x/cwerrors module.


## Consequences

### Positive
1. Allows the cosmwasm smart contracts on Archway to create ICA on other cosmos chains

### Negative
1. Since the ica packets are executed by the relayers, the ica account holders would not have to pay gas fees on the counterparty chain. This could lead to potential abuse by spamming txs for counterparty chain.
2. ICA functionality would still be limited by the counterparty chain and all the Msgs that they accept by their ica host module

