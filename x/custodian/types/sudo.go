package types

import (
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

type OpenAckDetails struct {
	PortID                string `json:"port_id"`
	ChannelID             string `json:"channel_id"`
	CounterpartyChannelID string `json:"counterparty_channel_id"`
	CounterpartyVersion   string `json:"counterparty_version"`
}

// MessageOnChanOpenAck is passed to a contract's sudo() entrypoint when an interchain
// account was successfully  registered.
type MessageOnChanOpenAck struct {
	OpenAck OpenAckDetails `json:"open_ack"`
}

// MessageSudoCallback is passed to a contract's sudo() entrypoint when an interchain
// transaction ended up with Success/Error or timed out.
type MessageSudoCallback struct {
	Response *ResponseSudoPayload `json:"response,omitempty"`
	Error    *ErrorSudoPayload    `json:"error,omitempty"`
	Timeout  *TimeoutPayload      `json:"timeout,omitempty"`
}

type ResponseSudoPayload struct {
	Request channeltypes.Packet `json:"request"`
	Data    []byte              `json:"data"` // Message data
}

type ErrorSudoPayload struct {
	Request channeltypes.Packet `json:"request"`
	Details string              `json:"details"`
}

type TimeoutPayload struct {
	Request channeltypes.Packet `json:"request"`
}
