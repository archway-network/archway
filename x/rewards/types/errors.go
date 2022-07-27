package types

import sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	DefaultCodespace    = ModuleName
	ErrInternal         = sdkErrors.Register(DefaultCodespace, 0, "internal error")         // smth went wrong
	ErrContractNotFound = sdkErrors.Register(DefaultCodespace, 1, "contract not found")     // contract info not found
	ErrUnauthorized     = sdkErrors.Register(DefaultCodespace, 2, "unauthorized operation") // contract ownership issue
)
