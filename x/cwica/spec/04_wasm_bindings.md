# Wasm Bindings

The module exposes custom bindings such that the contracts can access the IBC acknowledgements of the ICA interactions.

The proto definitions can be found [here](../../../proto/archway/cwica/v1/sudo.proto).

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

This is handeled by the [x/cwerrors](../../cwerrors/spec/README.md) module.

## Transaction Exectuion - Timeout

This is handeled by the [x/cwerrors](../../cwerrors/spec/README.md) module.

Please note that packet timeouts cause the ibc channel to be closed. The channel can be reopened again by registering the ica account again using [MsgRegisterInterchainAccount](../../../proto/archway/cwica/v1/tx.proto)
