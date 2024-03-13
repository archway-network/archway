# Messages

Section describes the processing of the module messages

## MsgUpdateParams

The module params can be updated via a governance proposal using the x/gov module. The proposal needs to include [MsgUpdateParams](../../../proto/archway/cwerrors/v1/tx.proto) message. All the parameters need to be provided when creating the msg.

```protobuf
message MsgUpdateParams {
    option (cosmos.msg.v1.signer) = "authority";
    // authority is the address that controls the module (defaults to x/gov unless overwritten).
    string authority = 1;
    // params defines the x/cwerrors parameters to update.
    //
    // NOTE: All parameters must be supplied.
    Params params = 2 [(gogoproto.nullable) = false, (gogoproto.jsontag) = "params,omitempty"];
}
```

On success: 
* Module `Params` are updated to the new values

This message is expected to fail if:
* The msg is sent by someone who is not the x/gov module
* The param values are invalid

## MsgSubscribeToError

A contract can be subscribed to errors by using the [MsgSubscribeToError](../../../proto/archway/cwerrors/v1/tx.proto) message.

```protobuf
message MsgSubscribeToError {
    option (cosmos.msg.v1.signer) = "sender";
    // sender is the address of who is registering the contarcts for callback on error
    string sender = 1;
    // contract is the address of the contract that will be called on error
    string contract_address = 2;
    // fee is the subscription fee for the feature (current no fee is charged for this feature)
    cosmos.base.v1beta1.Coin fee = 3 [ (gogoproto.nullable) = false ];
}
```

On success
* A subscription is created valid for the duration as specified in the module params.
* The subscription fees are sent to the fee collector
* In case a subscription already exists, it is extended.

This message is expected to fail if:
* The sender address and contract address are not valid addresses
* There is no contract with given address
* The sender is not authorized to subscribe - the sender is not the contract owner/admin or the contract itself
* The user does not send enough funds or doesnt have enough funds