package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (m *GenesisState) Validate() error {
	set := make(map[string]struct{}, len(m.GrantingContracts))
	for i, addr := range m.GrantingContracts {
		_, isDuplicate := set[addr]
		if isDuplicate {
			return fmt.Errorf("duplicate granting contract at index %d: %s", i, addr)
		}
		_, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return fmt.Errorf("invalid bech32 address of granting contract %d %s: %w", i, addr, err)
		}
		set[addr] = struct{}{}
	}
	return nil
}
