package types

import sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	DefaultCodespace                = ModuleName
	ErrInternal                     = sdkErrors.Register(DefaultCodespace, 0, "internal error") // smth went wrong
	ErrInvalidInflation             = sdkErrors.Register(DefaultCodespace, 1, "invalid inflation percentage")
	ErrInvalidTimestamp             = sdkErrors.Register(DefaultCodespace, 2, "invalid last block info timestamp")
	ErrInvalidBondedRatio           = sdkErrors.Register(DefaultCodespace, 3, "invalid bonded ratio")
	ErrInvalidInflationChange       = sdkErrors.Register(DefaultCodespace, 4, "invalid inflation change ratio")
	ErrInvalidMaxBlockDuration      = sdkErrors.Register(DefaultCodespace, 5, "invalid max block duration")
	ErrInvalidInflationRecipient    = sdkErrors.Register(DefaultCodespace, 6, "invalid inflation recipient")
	ErrInvalidInflationDistribution = sdkErrors.Register(DefaultCodespace, 7, "invalid inflation distribution")
)
