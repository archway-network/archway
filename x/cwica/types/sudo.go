package types

import (
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

// Sudopayload is the payload for the sudo call sent by the cwica module on IBC actions
type SudoPayload struct {
	ICA   *MessageICASuccess `json:"ica,omitempty"`
	Error *SudoErrorMsg      `json:"error,omitempty"`
}

// MessageICASuccess is the success message for the sudo call
type MessageICASuccess struct {
	AccountRegistered *AccountRegistered `json:"account_registered,omitempty"`
	AccountClosed     *ChannelClosed     `json:"account_closed,omitempty"`
	TxExecuted        *ICATxResponse     `json:"tx_executed,omitempty"`
}

// MessageICAError is the error message for the sudo call
type MessageICAError struct {
	Error *SudoErrorMsg `json:"error,omitempty"`
}

// AccountRegistered is contains the address of the registered account on the counterparty chain
type AccountRegistered struct {
	CounterpartyAddress string `json:"counterparty_address"`
}

type ChannelClosed struct {
	PortID    string `json:"port_id"`
	ChannelID string `json:"channel_id"`
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
