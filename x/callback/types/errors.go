package types

import errorsmod "cosmossdk.io/errors"

var (
	DefaultCodespace = ModuleName
	ErrInternal      = errorsmod.Register(DefaultCodespace, 0, "internal error") // smth went wrong
)
