package types

import sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	DefaultCodespace    = ModuleName
	ErrInternal         = sdkErrors.Register(DefaultCodespace, 0, "internal error") // smth went wrong
	ErrInvalidInflation = sdkErrors.Register(DefaultCodespace, 1, "invalid inflation percentage")
	ErrInvalidTimestamp = sdkErrors.Register(DefaultCodespace, 2, "invalid last block info timestamp")
)
