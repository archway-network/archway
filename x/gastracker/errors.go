package gastracker

import (
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	DefaultCodespace = ModuleName

	ErrBlockTrackingDataNotFound = sdkErrors.Register(DefaultCodespace, 1, "Block tracking data not found")

	ErrTxTrackingDataNotFound = sdkErrors.Register(DefaultCodespace, 2, "Tx tracking data not found")

	ErrContractInstanceMetadataNotFound = sdkErrors.Register(DefaultCodespace, 3, "Contract instance metadata not found")

	ErrCurrentBlockTrackingDataAlreadyExists = sdkErrors.Register(DefaultCodespace, 4, "Current block tracking data already exists")

	ErrRewardEntryNotFound = sdkErrors.Register(DefaultCodespace, 5, "Reward entry not found")

	ErrInvalidInitRequest1 = sdkErrors.Register(DefaultCodespace, 6, "Invalid instantiation request, you cannot have both gas rebate and premium charge true")

	ErrInvalidInitRequest2 = sdkErrors.Register(DefaultCodespace, 7, "Invalid instantiation request, premium percentage is out of range")

	ErrContractInfoNotFound = sdkErrors.Register(DefaultCodespace, 8, "Contract info not found")

	ErrNoPermissionToSetMetadata = sdkErrors.Register(DefaultCodespace, 9, "Sender does not have permission to set metadata")

	ErrInvalidSetContractMetadataRequest = sdkErrors.Register(DefaultCodespace, 10, "Invalid request to set metadata")
)
