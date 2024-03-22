package types

import (
	"cosmossdk.io/errors"
	cwerrortypes "github.com/archway-network/archway/x/cwerrors/types"
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

// NewSudoError creates a new sudo error instance to pass on to the errors module
func NewSudoError(errorCode ModuleErrors, contractAddr string, inputPayload string, errMsg string) cwerrortypes.SudoError {
	return cwerrortypes.SudoError{
		ModuleName:      ModuleName,
		ErrorCode:       int32(errorCode),
		ContractAddress: contractAddr,
		InputPayload:    inputPayload,
		ErrorMessage:    errMsg,
	}
}
