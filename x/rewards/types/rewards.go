package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"sigs.k8s.io/yaml"
)

// Validate performs object fields validation.
func (m ContractMetadata) Validate() error {
	if _, err := sdk.AccAddressFromBech32(m.OwnerAddress); err != nil {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "invalid owner address")
	}
	if _, err := sdk.AccAddressFromBech32(m.RewardsAddress); err != nil {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "invalid rewards address")
	}

	return nil
}

// String implements the fmt.Stringer interface.
func (m ContractMetadata) String() string {
	bz, _ := yaml.Marshal(m)
	return string(bz)
}
