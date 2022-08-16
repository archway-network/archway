package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"sigs.k8s.io/yaml"
)

// HasOwnerAddress returns true if the rewards address is set.
func (m ContractMetadata) HasOwnerAddress() bool {
	return m.OwnerAddress != ""
}

// HasRewardsAddress returns true if the rewards address is set.
func (m ContractMetadata) HasRewardsAddress() bool {
	return m.RewardsAddress != ""
}

// MustGetContractAddress returns the contract address.
// CONTRACT: panics in case of an error.
func (m ContractMetadata) MustGetContractAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.ContractAddress)
	if err != nil {
		panic(fmt.Errorf("parsing contract address: %w", err))
	}

	return addr
}

// MustGetRewardsAddress returns the rewards address.
// CONTRACT: panics in case of an error.
func (m ContractMetadata) MustGetRewardsAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.RewardsAddress)
	if err != nil {
		panic(fmt.Errorf("parsing rewards address (%s): %s", m.RewardsAddress, err))
	}

	return addr
}

// Validate performs object fields validation.
// genesisValidation flag perform strict validation of the genesis state (some field must be set).
func (m ContractMetadata) Validate(genesisValidation bool) error {
	if _, err := sdk.AccAddressFromBech32(m.ContractAddress); err != nil {
		return sdkErrors.Wrapf(sdkErrors.ErrInvalidAddress, "invalid contract address: %v", err)
	}

	if genesisValidation || m.OwnerAddress != "" {
		if _, err := sdk.AccAddressFromBech32(m.OwnerAddress); err != nil {
			return sdkErrors.Wrapf(sdkErrors.ErrInvalidAddress, "invalid owner address: %v", err)
		}
	}

	if m.RewardsAddress != "" {
		if _, err := sdk.AccAddressFromBech32(m.RewardsAddress); err != nil {
			return sdkErrors.Wrapf(sdkErrors.ErrInvalidAddress, "invalid rewards address: %v", err)
		}
	}

	return nil
}

// String implements the fmt.Stringer interface.
func (m ContractMetadata) String() string {
	bz, _ := yaml.Marshal(m)
	return string(bz)
}
