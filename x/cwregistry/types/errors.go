package types

import (
	"cosmossdk.io/errors"
)

// x/cwregistry module sentinel errors
var (
	ErrNotContract = errors.Register(ModuleName, 1103, "not a contract")
)
