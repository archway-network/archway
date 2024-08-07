package types

import (
	"fmt"
)

var (
	DefaultMsgSendTxMaxMessages = uint64(5)
)

// NewParams creates a new Params instance
func NewParams(msgSendTxMaxMessages uint64) Params {
	return Params{
		MsgSendTxMaxMessages: msgSendTxMaxMessages,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultMsgSendTxMaxMessages)
}

// Validate validates the set of params
func (p Params) Validate() error {
	return validateMsgSendTxMaxMessages(p.GetMsgSendTxMaxMessages())
}

func validateMsgSendTxMaxMessages(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("MsgSendTxMaxMessages must be greater than zero")
	}

	return nil
}
