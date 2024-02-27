package types

import (
	"cosmossdk.io/errors"
)

// x/cwica module sentinel errors
var (
	ErrInvalidICAOwner           = errors.Register(ModuleName, 1100, "invalid interchain account interchainAccountID")
	ErrInvalidAccountAddress     = errors.Register(ModuleName, 1101, "invalid account address")
	ErrInterchainAccountNotFound = errors.Register(ModuleName, 1102, "interchain account not found")
	ErrNotContract               = errors.Register(ModuleName, 1103, "not a contract")
	ErrEmptyConnectionID         = errors.Register(ModuleName, 1105, "empty connection id")
	ErrNoMessages                = errors.Register(ModuleName, 1106, "no messages provided")
	ErrInvalidTimeout            = errors.Register(ModuleName, 1107, "invalid timeout")
	ErrLongInterchainAccountID   = errors.Register(ModuleName, 1108, "interchain account id is too long")
)

type SudoErrorMsg struct {
	ModuleName   string       `json:"module_name"`
	ErrorCode    ModuleErrors `json:"error_code"`
	InputPayload string       `json:"input_payload"`
	ErrorMessage string       `json:"error_message"`
}

// SudoEssosMsg constructor
func NewSudoErrorMsg(err SudoError) *SudoErrorMsg {
	return &SudoErrorMsg{
		ModuleName:   ModuleName,
		ErrorCode:    err.GetErrorCode(),
		InputPayload: err.Payload,
		ErrorMessage: err.ErrorMsg,
	}
}
