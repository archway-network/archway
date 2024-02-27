# Wasm Bindings

The module exposes custom bindings such that the contracts can access the IBC acknowledgements of the ICA interactions.

```go
// Sudopayload is the payload for the sudo call sent by the cwica module on IBC actions
type SudoPayload struct {
    // ICA is the endpoint name at the contract which is called
	ICA   *MessageICASuccess `json:"ica,omitempty"`
}

// MessageICASuccess is the success message for the sudo call
type MessageICASuccess struct {
    // AccountRegistered is populated when a new interchain account has been successfully registered
	AccountRegistered *OpenAckDetails `json:"account_registered,omitempty"`
    // TxExecuted is populated when an ica transactin has been successfully executed
	TxExecuted        *ICATxResponse  `json:"tx_executed,omitempty"`
}

// OpenAckDetails is the details of the open ack message - when an interchain account is registered
type OpenAckDetails struct {
	PortID                string `json:"port_id"`
	ChannelID             string `json:"channel_id"`
	CounterpartyChannelID string `json:"counterparty_channel_id"`
	CounterpartyVersion   string `json:"counterparty_version"`
}

// ICATxResponse is the response message after the execute of the ICA tx
type ICATxResponse struct {
	Packet channeltypes.Packet `json:"packet"`
	Data   []byte              `json:"data"` // Message response
}
```