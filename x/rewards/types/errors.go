package types

import sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	DefaultCodespace           = ModuleName
	ErrInternal                = sdkErrors.Register(DefaultCodespace, 0, "internal error")         // smth went wrong
	ErrContractNotFound        = sdkErrors.Register(DefaultCodespace, 1, "contract not found")     // contract info not found
	ErrMetadataNotFound        = sdkErrors.Register(DefaultCodespace, 2, "metadata not found")     // contract metadata not found
	ErrUnauthorized            = sdkErrors.Register(DefaultCodespace, 3, "unauthorized operation") // contract ownership issue
	ErrInvalidRequest          = sdkErrors.Register(DefaultCodespace, 4, "invalid request")        // request parsing issue
	ErrContractFlatFeeNotFound = sdkErrors.Register(DefaultCodespace, 5, "flatfee not found")      // contract flatfee not found
)
