package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/archway-network/archway/pkg"
)

// HasRewards returns true if the block rewards have been set.
func (m BlockRewards) HasRewards() bool {
	return !m.InflationRewards.IsZero()
}

// HasGasLimit returns true if the gas limit has been set.
func (m BlockRewards) HasGasLimit() bool {
	return m.MaxGas > 0
}

// Validate performs object fields validation.
func (m BlockRewards) Validate() error {
	if m.Height < 0 {
		return fmt.Errorf("height: must be GTE 0")
	}

	if !pkg.CoinIsZero(m.InflationRewards) {
		if err := pkg.ValidateCoin(m.InflationRewards); err != nil {
			return fmt.Errorf("inflationRewards: %w", err)
		}
	}

	return nil
}

// HasRewards returns true if the transaction rewards have been set.
func (m TxRewards) HasRewards() bool {
	return !sdk.Coins(m.FeeRewards).IsZero()
}

// Validate performs object fields validation.
func (m TxRewards) Validate() error {
	if m.TxId <= 0 {
		return fmt.Errorf("txId: must be GT 0")
	}

	if m.Height < 0 {
		return fmt.Errorf("height: must be GTE 0")
	}

	for i, coin := range m.FeeRewards {
		if pkg.CoinIsZero(coin) {
			return fmt.Errorf("feeRewards [%d]: must be non-zero", i)
		}

		if err := pkg.ValidateCoin(coin); err != nil {
			return fmt.Errorf("feeRewards [%d]: %w", i, err)
		}
	}

	return nil
}

// MustGetRewardsAddress returns the rewards address.
// CONTRACT: panics in case of an error.
func (m RewardsRecord) MustGetRewardsAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.RewardsAddress)
	if err != nil {
		panic(fmt.Errorf("parsing rewardsRecord rewardsAddress: %w", err))
	}
	return addr
}

// Validate performs object fields validation.
func (m RewardsRecord) Validate() error {
	if m.Id <= 0 {
		return fmt.Errorf("id: must be GT 0")
	}

	if _, err := sdk.AccAddressFromBech32(m.RewardsAddress); err != nil {
		return fmt.Errorf("rewardsAddress: %w", err)
	}

	for i, coin := range m.Rewards {
		if err := pkg.ValidateCoin(coin); err != nil {
			return fmt.Errorf("rewards [%d]: %w", i, err)
		}
	}

	if m.CalculatedHeight < 0 {
		return fmt.Errorf("calculatedHeight: must be GTE 0")
	}

	if m.CalculatedTime.IsZero() {
		return fmt.Errorf("calculatedTime: must be non-zero")
	}

	return nil
}

// Validate performs object fields validation.
func (m FlatFee) Validate() error {
	if _, err := sdk.AccAddressFromBech32(m.ContractAddress); err != nil {
		return errorsmod.Wrapf(sdkErrors.ErrInvalidAddress, "invalid contract address: %v", err)
	}

	if err := pkg.ValidateCoin(m.FlatFee); err != nil {
		return errorsmod.Wrapf(sdkErrors.ErrInvalidCoins, "invalid flat fee coin: %v", err)
	}

	return nil
}

// MustGetContractAddress returns the contract address.
// CONTRACT: panics in case of an error.
func (m FlatFee) MustGetContractAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.ContractAddress)
	if err != nil {
		panic(fmt.Errorf("parsing contract address: %w", err))
	}

	return addr
}
