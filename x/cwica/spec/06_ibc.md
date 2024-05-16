# IBC Handlers

Section describes the processing of the various IBC events

## OnChanOpenInit

There is no custom logic executed under this handler.

## OnChanOpenTry

This handler is not implemented for controller, only host.

## OnChanOpenAck

This handler is executed when a channel is successfully created between the chain and the counterparty chain. 

On success ack: 
* The handler parses the counterparty version details and extracts the counterparty address and executes a Sudo call to the contract with the details

## OnChanOpenConfirm

This handler is not implemented for controller, only host.

## OnChanCloseInit

This handler is not implemented for controller, only host.

## OnChanCloseConfirm

This handler is currently not implemented by the module

## OnRecvPacket

This handler is not implemented for controller, only host.

## OnAcknowledgementPacket

This handler is executed when an ibc acknowledgement packet is received for a MsgSendTx operation.

On success:
* The handler formats the ibc packet and acknowledgement result into a sudo payload and executes a Sudo call to the contract with the details.

On failure:
* The handler formats the ibc packet and error into a Sudo Error with error code [ERR_EXEC_FAILURE](05_errors.md) and sends it to the x/cwerrors module

## OnTimeoutPacket

This handler is executed when an ibc packet times out. It formats the ibc packet int oa Sudo Error with error code  [ERR_PACKET_TIMEOUT](05_errors.md) and sends it to the x/cwerrors module