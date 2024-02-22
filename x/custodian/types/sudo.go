package types

import (
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

// Sudopayload is the payload for the sudo call sent by the custodian module on IBC actions
type SudoPayload struct {
	Custodian *MessageCustodianSuccess `json:"custodian,omitempty"`
	Error     *SudoErrorMsg            `json:"error,omitempty"`
}

// MessageCustodianSuccess is the success message for the sudo call
type MessageCustodianSuccess struct {
	AccountRegistered *OpenAckDetails `json:"account_registered,omitempty"`
	TxExecuted        *ICATxResponse  `json:"tx_executed,omitempty"`
}

// MessageCustodianError is the error message for the sudo call
type MessageCustodianError struct {
	Error *SudoErrorMsg `json:"error,omitempty"`
	// Failure *ICATxError   `json:"failure,omitempty"`
	// Timeout *ICATxTimeout `json:"timeout,omitempty"`
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

// ICATxError is the error message after the execute of the ICA tx
type ICATxError struct {
	Packet  channeltypes.Packet `json:"packet"`
	Details string              `json:"details"`
}

// ICATxTimeout is the timeout message after the execute of the ICA tx
type ICATxTimeout struct {
	Packet channeltypes.Packet `json:"packet"`
}
