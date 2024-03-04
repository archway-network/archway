package types

import (
	"cosmossdk.io/errors"
)

// x/cwica module sentinel errors
var (
	ErrInvalidAccountAddress            = errors.Register(ModuleName, 1101, "invalid account address")
	ErrInterchainAccountNotFound        = errors.Register(ModuleName, 1102, "interchain account not found")
	ErrNotContract                      = errors.Register(ModuleName, 1103, "not a contract")
	ErrEmptyConnectionID                = errors.Register(ModuleName, 1104, "empty connection id")
	ErrCounterpartyConnectionNotFoundID = errors.Register(ModuleName, 1105, "counterparty connection id not found")
	ErrNoMessages                       = errors.Register(ModuleName, 1106, "no messages provided")
	ErrInvalidTimeout                   = errors.Register(ModuleName, 1107, "invalid timeout")
)

// SudoEssosMsg constructor
func NewSudoError(err SudoError) *SudoError {
	err.ModuleName = ModuleName
	return &err
}
