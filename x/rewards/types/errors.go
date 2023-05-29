package types

import sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	DefaultCodespace           = ModuleName
	ErrInternal                = sdkErrors.Register(DefaultCodespace, 2, "internal error")         // internal error
	ErrContractNotFound        = sdkErrors.Register(DefaultCodespace, 3, "contract not found")     // contract info not found
	ErrMetadataNotFound        = sdkErrors.Register(DefaultCodespace, 4, "metadata not found")     // contract metadata not found
	ErrUnauthorized            = sdkErrors.Register(DefaultCodespace, 5, "unauthorized operation") // contract ownership issue
	ErrInvalidRequest          = sdkErrors.Register(DefaultCodespace, 6, "invalid request")        // request parsing issue
	ErrContractFlatFeeNotFound = sdkErrors.Register(DefaultCodespace, 7, "flatfee not found")      // contract flatfee not found
)
