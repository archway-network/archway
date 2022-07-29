package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"sigs.k8s.io/yaml"
)

// HasRewards returns true if the block rewards have been set.
func (m BlockRewards) HasRewards() bool {
	return !m.InflationRewards.IsZero()
}

// Validate performs object fields validation.
func (m BlockRewards) Validate() error {
	if m.Height < 0 {
		return fmt.Errorf("height: must be GTE 0")
	}

	if m.InflationRewards.Denom != "" {
		if err := sdk.ValidateDenom(m.InflationRewards.Denom); err != nil {
			return fmt.Errorf("inflationRewards: denom: %w", err)
		}
	}
	if m.InflationRewards.Amount.IsNegative() {
		return fmt.Errorf("inflationRewards: amount: is negative")
	}

	if m.MaxGas <= 0 {
		return fmt.Errorf("maxGas: must be GT 0")
	}

	return nil
}

// String implements the fmt.Stringer interface.
func (m BlockRewards) String() string {
	bz, _ := yaml.Marshal(m)
	return string(bz)
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
		if err := sdk.ValidateDenom(coin.Denom); err != nil {
			return fmt.Errorf("feeRewards [%d]: denom: %w", i, err)
		}
		if coin.Amount.IsNegative() {
			return fmt.Errorf("feeRewards [%d]: amount: is negative", i)
		}
	}

	return nil
}

// String implements the fmt.Stringer interface.
func (m TxRewards) String() string {
	bz, _ := yaml.Marshal(m)
	return string(bz)
}

// String implements the fmt.Stringer interface.
func (m BlockTracking) String() string {
	bz, _ := yaml.Marshal(m)
	return string(bz)
}
