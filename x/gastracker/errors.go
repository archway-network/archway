package gastracker

import (
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	DefaultCodespace                         = ModuleName
	ErrInternal                              = sdkErrors.Register(DefaultCodespace, 0, "internal error")
	ErrBlockTrackingDataNotFound             = sdkErrors.Register(DefaultCodespace, 1, "block tracking data not found")
	ErrTxTrackingDataNotFound                = sdkErrors.Register(DefaultCodespace, 2, "tx tracking data not found")
	ErrContractInstanceMetadataNotFound      = sdkErrors.Register(DefaultCodespace, 3, "contract instance metadata not found")
	ErrCurrentBlockTrackingDataAlreadyExists = sdkErrors.Register(DefaultCodespace, 4, "current block tracking data already exists")
	ErrRewardEntryNotFound                   = sdkErrors.Register(DefaultCodespace, 5, "reward entry not found")
	ErrInvalidInitRequest1                   = sdkErrors.Register(DefaultCodespace, 6, "invalid instantiation request, you cannot have both gas rebate and premium charge true")
	ErrInvalidInitRequest2                   = sdkErrors.Register(DefaultCodespace, 7, "invalid instantiation request, premium percentage is out of range")
	ErrContractInfoNotFound                  = sdkErrors.Register(DefaultCodespace, 8, "contract info not found")
	ErrNoPermissionToSetMetadata             = sdkErrors.Register(DefaultCodespace, 9, "sender does not have permission to set metadata")
	ErrInvalidSetContractMetadataRequest     = sdkErrors.Register(DefaultCodespace, 10, "invalid request to set metadata")
	ErrDappInflationaryRewardRecordNotFound  = sdkErrors.Register(DefaultCodespace, 11, "inflationary rewards record not found")
	ErrInvalidRequest                        = sdkErrors.Register(DefaultCodespace, 12, "invalid request")
)
