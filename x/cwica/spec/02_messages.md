# Messages

Section describes the processing of the module messages

## MsgUpdateParams

The module params can be updated via a governance proposal using the x/gov module. The proposal needs to include [MsgUpdateParams](../../../proto/archway/cwica/v1/tx.proto#L70) message. All the parameters need to be provided when creating the msg.

```proto
message MsgUpdateParams {
  option (amino.name) = "cwica/MsgUpdateParams";
  option (cosmos.msg.v1.signer) = "authority";

  // authority is the address of the authority that is allowed to update the
  // cwica module parameters.
  string authority = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];

  // params deines the module parmeters to update
  // NOTE: All parameters must be supplied.
  Params params = 2
      [ (gogoproto.nullable) = false, (amino.dont_omitempty) = true ];
}
```

On success: 
* Module `Params` are updated to the new values

This message is expected to fail if:
* The msg is sent by someone who is not the x/gov module
* The param values are invalid

## MsgRegisterInterchainAccount

A new interchain account can be registered by using the [MsgRegisterInterchainAccount](../../../proto/archway/cwica/v1/tx.proto#L29) message.

```proto
message MsgRegisterInterchainAccount {
  option (gogoproto.equal) = false;
  option (gogoproto.goproto_getters) = false;
  // contract_address is the address of the contract who wants to register an ica account on
  // the counterparty chain
  string contract_address = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
  // connection_id is the connection id between the two chains
  string connection_id = 2 [ (gogoproto.moretags) = "yaml:\"connection_id\"" ];
}
```

On Success
* An ibc packet is created which will attempt to create a new channel on the given connection

This message is expected to fail if:
* The sender/ From Address is not a cosmwasm smart contract
* The connection id is non existent

## MsgSendTx

A collection of transactions can be submitted to be executed as a transaction from the ICA account on a counterparty chain by using the [MsgSendTx](../../../proto/archway/cwica/v1/tx.proto#L43)

```proto
message MsgSendTx {
  option (gogoproto.equal) = false;
  option (gogoproto.goproto_getters) = false;
  // contract_address is the address of the contract who wants to submit a transaction to the
  // counterparty chain
  string contract_address = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
  // connection_id is the connection id between the two chains
  string connection_id = 2;
  // msgs are the messages to be submitted to the counterparty chain
  repeated google.protobuf.Any msgs = 3;
  // memo is the memo to be included in the packet
  string memo = 4;
  // timeout in seconds after which the packet times out
  uint64 timeout = 5;
}
```

On Success
* An ibc packet is created which will be executed on the counterparty chain

This mesasge is expected to fail if:
* There are no msgs to be executed
* Sender is not a smart contract
* There are more than allowed number of msgs (see [Params](../../../proto/archway/cwica/v1/params.proto))
* There is no active channel for given connection and port 