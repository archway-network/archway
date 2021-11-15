package types

import (
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	DefaultCodespace = ModuleName

	ErrBlockTrackingDataNotFound = sdkErrors.Register(DefaultCodespace, 1, "Block tracking data not found")

	ErrTxTrackingDataNotFound = sdkErrors.Register(DefaultCodespace, 2, "Tx tracking data not found")

	ErrContractInstanceMetadataNotFound = sdkErrors.Register(DefaultCodespace, 3, "Contract instance metadata not found")

	ErrRewardEntryNotFound = sdkErrors.Register(DefaultCodespace, 4, "Reward entry not found")

	ErrInvalidInitRequest1 = sdkErrors.Register(DefaultCodespace, 5, "Invalid instantiation request, you cannot have both gas rebate and premium charge true")

	ErrInvalidInitRequest2 = sdkErrors.Register(DefaultCodespace, 6, "Invalid instantiation request, premium percentage is out of range")
)
