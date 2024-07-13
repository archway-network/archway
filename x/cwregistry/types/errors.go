package types

import (
	"cosmossdk.io/errors"
)

// x/cwregistry module sentinel errors
var (
	ErrUnauthorized   = errors.Register(ModuleName, 1100, "sender unauthorized to set the metadata")
	ErrNoSuchCode     = errors.Register(ModuleName, 1104, "code with given id does not exist")
	ErrNoSuchContract = errors.Register(ModuleName, 1105, "contract with given address does not exist")
	ErrSchemaTooLarge = errors.Register(ModuleName, 1106, "schema cannot be larger than 255 bytes")
	// Snapshot related errors
	ErrInvalidSnapshotPayload = errors.Register(ModuleName, 1107, "invalid snapshot payload")
)
