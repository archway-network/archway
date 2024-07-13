# Messages

Section describes the processing of the module messages


## MsgRegisterCode

Contract code metadata can be registered by the [MsgRegisterCode](../../../proto/archway/cwregistry/v1/tx.proto#L21) message.

On success:
* The provided code metadata is stored or updated

This message is expected to fail if:
* Sender address is invalid
* Given Code ID is nto valid
* The sender is not the address who uploaded the contract binary
* Any of the fields are longer than 255 characters.
* The source code repository is now a valid URL
 