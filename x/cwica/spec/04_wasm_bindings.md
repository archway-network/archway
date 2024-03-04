# Wasm Bindings

The module exposes custom bindings such that the contracts can access the IBC acknowledgements of the ICA interactions.

The proto definitions can be found [here](../../../proto/archway/cwica/v1/sudo.proto)

## Account Creation Callback

After successfull interchain account creation on the counterparty chain, the sudo entrypoint is called with the following json

```jsonc
{
	"ica" : {
		"account_registered": {
			"counterparty_address": "cosmos1j58ehredzltrfp0e5nkyat9w549rumah8r3fln" // the address on the counterparty chain
		}
	}
}
```

## Transaction Execution - Success

After successful interchain account transaction execution on the counterparty chain, the sudo entrypoint is called with the following json

```jsonc
{
	"ica": {
		"tx_executed": {
			"data" : [], // byte array of the response of the tx execution
			"packet": {  
				// the ibc packet info. Find more details in channeltypes.Packet in ibc-go proto defs
				// https://github.com/cosmos/ibc-go/blob/v7.3.2/proto/ibc/core/channel/v1/channel.proto
			}
		}
	}
}
```

## Transaction Execution - Failed 

If the interachain account transation failed on the counterparty chain, the sudo entrypoint will be called with the following json

```jsonc
{
	"error": {
		"module_name": "cwica",
		"error_code": 2, // More details, look at archway/cwica/v1/errors.proto
		"input_payload": "", // ibc packet info
		"error_message": "" // any relevant error message sent by the counterparty chain
	}
}
```

## Transaction Exectuion - Timeout

In case the ibc packet timed out ([more info on packet timeouts](https://ibc.cosmos.network/v7/ibc/overview?_highlight=timeout#receipts-and-timeouts)),  the sudo entrypoint will be called with the following json

```jsonc
{
	"error": {
		"module_name": "cwica",
		"error_code": 1, // More details, look at archway/cwica/v1/errors.proto
		"input_payload": "", // ibc packet info
		"error_message": "IBC packet timeout" 
	}
}
```

Please note that packet timeouts cause the ibc channel to be closed. The channel can be reopened again by registering the ica account again using [MsgRegisterInterchainAccount](../../../proto/archway/cwica/v1/tx.proto)

## Error Codes

The error codes used by the module are

```protobuf
enum ModuleErrors {
  ERR_UNKNOWN = 0;
  ERR_PACKET_TIMEOUT = 1; // When the ibc packet timesout
  ERR_EXEC_FAILURE = 2; // When tx execution fails on counterparty chain
}
```
