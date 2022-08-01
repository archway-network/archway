package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"sigs.k8s.io/yaml"

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
		if pkg.CoinIsZero(coin) {
			return fmt.Errorf("feeRewards [%d]: must be non-zero", i)
		}

		if err := pkg.ValidateCoin(coin); err != nil {
			return fmt.Errorf("feeRewards [%d]: %w", i, err)
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
