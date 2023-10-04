package types

import errorsmod "cosmossdk.io/errors"

var (
	DefaultCodespace           = ModuleName
	ErrInternal                = errorsmod.Register(DefaultCodespace, 2, "internal error")         // internal error
	ErrContractNotFound        = errorsmod.Register(DefaultCodespace, 3, "contract not found")     // contract info not found
	ErrMetadataNotFound        = errorsmod.Register(DefaultCodespace, 4, "metadata not found")     // contract metadata not found
	ErrUnauthorized            = errorsmod.Register(DefaultCodespace, 5, "unauthorized operation") // contract ownership issue
	ErrInvalidRequest          = errorsmod.Register(DefaultCodespace, 6, "invalid request")        // request parsing issue
	ErrContractFlatFeeNotFound = errorsmod.Register(DefaultCodespace, 7, "flatfee not found")      // contract flatfee not found
)
