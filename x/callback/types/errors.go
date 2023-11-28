package types

import errorsmod "cosmossdk.io/errors"

var (
	DefaultCodespace             = ModuleName
	ErrContractNotFound          = errorsmod.Register(DefaultCodespace, 2, "contract with given address not found")
	ErrCallbackJobIDExists       = errorsmod.Register(DefaultCodespace, 3, "callback with given job id already exists for given height")
	ErrCallbackHeightNotinFuture = errorsmod.Register(DefaultCodespace, 4, "callback request height is not in the future")
	ErrUnauthorized              = errorsmod.Register(DefaultCodespace, 5, "sender not authorized to register callback")
	ErrCallbackNotFound          = errorsmod.Register(DefaultCodespace, 6, "callback with given job id does not exist for given height")
	ErrInsufficientFees          = errorsmod.Register(DefaultCodespace, 7, "insufficient fees to register callback")
)
