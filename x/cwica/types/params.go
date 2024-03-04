package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	KeyMsgSendTxMaxMessages     = []byte("MsgSendTxMaxMessages")
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

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(
			KeyMsgSendTxMaxMessages,
			&p.MsgSendTxMaxMessages,
			validateMsgSendTxMaxMessages,
		),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	return validateMsgSendTxMaxMessages(p.GetMsgSendTxMaxMessages())
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
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
