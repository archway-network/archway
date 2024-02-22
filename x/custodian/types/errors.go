package types

import (
	"cosmossdk.io/errors"
)

// x/custodian module sentinel errors
var (
	ErrInvalidICAOwner           = errors.Register(ModuleName, 1100, "invalid interchain account interchainAccountID")
	ErrInvalidAccountAddress     = errors.Register(ModuleName, 1101, "invalid account address")
	ErrInterchainAccountNotFound = errors.Register(ModuleName, 1102, "interchain account not found")
	ErrNotContract               = errors.Register(ModuleName, 1103, "not a contract")
	ErrEmptyInterchainAccountID  = errors.Register(ModuleName, 1104, "empty interchain account id")
	ErrEmptyConnectionID         = errors.Register(ModuleName, 1105, "empty connection id")
	ErrNoMessages                = errors.Register(ModuleName, 1106, "no messages provided")
	ErrInvalidTimeout            = errors.Register(ModuleName, 1107, "invalid timeout")
	ErrLongInterchainAccountID   = errors.Register(ModuleName, 1109, "interchain account id is too long")
)

type SudoErrorMsg struct {
	ModuleName   string       `json:"module_name"`
	ErrorCode    ModuleErrors `json:"error_code"`
	Payload      string       `json:"payload"`
	ErrorMessage string       `json:"error_message"`
}

// SudoEssosMsg constructor
func NewSudoErrorMsg(err SudoError) *SudoErrorMsg {
	return &SudoErrorMsg{
		ModuleName:   ModuleName,
		ErrorCode:    err.GetErrorCode(),
		Payload:      err.Payload,
		ErrorMessage: err.ErrorMsg,
	}
}
