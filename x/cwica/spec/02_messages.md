# Messages

Section describes the processing of the module messages

## MsgUpdateParams

The module params can be updated via a governance proposal using the x/gov module. The proposal needs to include [MsgUpdateParams](../../../proto/archway/cwica/v1/tx.proto#L76) message. All the parameters need to be provided when creating the msg.

On success: 
* Module `Params` are updated to the new values

This message is expected to fail if:
* The msg is sent by someone who is not the x/gov module
* The param values are invalid

## MsgRegisterInterchainAccount

A new interchain account can be registered by using the [MsgRegisterInterchainAccount](../../../proto/archway/cwica/v1/tx.proto#L29) message.

On Success
* An ibc packet is created which will attempt to create a new channel on the given connection

This message is expected to fail if:
* The sender/ From Address is not a cosmwasm smart contract

## MsgSubmitTx

A collection of transactions can be submitted to be executed as a transaction from the ICA account on a counterparty chain by using the [MsgSubmitTx](../../../proto/archway/cwica/v1/tx.proto#L47)

On Success
* An ibc packet is created which will be executed on the ocunterparty chain

This mesasge is expected to fail if:
* There are no msgs to be executed
* Sender is not a smart contract
* There are more than allowed number of msgs (see [Params](../../../proto/archway/cwica/v1/params.proto))
* There is no active channel for given connection and port 