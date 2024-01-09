package types

import (
	"fmt"

	"github.com/CosmWasm/wasmd/x/wasm/types"
	wasmVmTypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

// SudoMsg defines the message sudo enum that is sent to the CosmWasm contract.
type SudoMsg struct {
	// CWGrant defines the enum variant of the grant message.
	CWGrant *CWGrant `json:"cw_grant"`
}

// CWGrant defines the CWGrant variant of the SudoMsg.
type CWGrant struct {
	// FeeRequested defines the amount of fees needed to cover TX expenses.
	FeeRequested wasmVmTypes.Coins `json:"fee_requested"`
	// Msgs defines the list of messages which we're trying to execute.
	Msgs []CWGrantMessage `json:"msgs"`
}

// CWGrantMessage defines the TX message requesting for a grant.
type CWGrantMessage struct {
	// Sender defines the sender of the message, populated
	// using msg.GetSigners()[0].
	Sender string `json:"sender"`
	// TypeUrl defines the type URL of the message without backlashes.
	TypeUrl string `json:"type_url"`
	// Msg defines the base64 encoded bytes of the message.
	// A combination of TypeUrl and Msg can be used to decode
	// into a concrete Rust/Go/Etc contract type.
	Msg []byte `json:"msg"`
}

func NewSudoMsg(cdc codec.BinaryCodec, requestedFees sdk.Coins, msgs []sdk.Msg) (*SudoMsg, error) {
	cwGrantMsgs, err := NewCWGrantMessages(cdc, msgs)
	if err != nil {
		return nil, err
	}
	wasmdRequestedFees := types.NewWasmCoins(requestedFees)
	return &SudoMsg{CWGrant: &CWGrant{
		FeeRequested: wasmdRequestedFees,
		Msgs:         cwGrantMsgs,
	}}, nil
}

func NewCWGrantMessages(cdc codec.BinaryCodec, msgs []sdk.Msg) ([]CWGrantMessage, error) {
	m := make([]CWGrantMessage, len(msgs))
	for i := range msgs {
		msg, err := NewCWGrantMessage(cdc, msgs[i])
		if err != nil {
			return nil, fmt.Errorf("unable to convert message at index %d, into a CWGrant", i)
		}
		m[i] = msg
	}
	return m, nil
}

func NewCWGrantMessage(cdc codec.BinaryCodec, msg sdk.Msg) (CWGrantMessage, error) {
	sender := msg.GetSigners()
	if len(sender) != 1 {
		return CWGrantMessage{}, fmt.Errorf("cw grants on multi signer messages are disallowed, got number of signers: %d", len(sender))
	}
	protoMarshaler, ok := msg.(codec.ProtoMarshaler)
	if !ok {
		return CWGrantMessage{}, fmt.Errorf("not a codec.ProtoMarshaler")
	}
	msgBytes, err := cdc.Marshal(protoMarshaler)
	if err != nil {
		return CWGrantMessage{}, err
	}

	return CWGrantMessage{
		Sender:  sender[0].String(),
		TypeUrl: proto.MessageName(msg),
		Msg:     msgBytes,
	}, nil
}
