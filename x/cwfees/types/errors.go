package types

import "cosmossdk.io/errors"

var (
	ErrNotAContract   = errors.New(ModuleName, 0, "not a cosmwasm contract")
	ErrAlreadyGranter = errors.New(ModuleName, 1, "provided contract is already a granter")
	ErrNotAGranter    = errors.New(ModuleName, 2, "provided contract is not a granter")
)
