package types

import errorsmod "cosmossdk.io/errors"

var (
	DefaultCodespace    = ModuleName
	ErrContractNotFound = errorsmod.Register(DefaultCodespace, 2, "contract with given address not found")
	ErrUnauthorized     = errorsmod.Register(DefaultCodespace, 3, "sender unauthorized to perform the action")
)
