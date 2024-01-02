package types

import "cosmossdk.io/errors"

var (
	ErrNotAContract   = errors.New(ModuleName, 0, "not a cosmwasm contract")
	ErrAlreadyGranter = errors.New(ModuleName, 1, "provided contract is already a granter")
)
