package types

// NewGenesisState creates a new GenesisState object.
func NewGenesisState(
	params Params,
	lbi LastBlockInfo,
) *GenesisState {
	return &GenesisState{
		Params:        params,
		LastBlockInfo: lbi,
	}
}

// DefaultGenesisState returns a default genesis state.
func DefaultGenesis() *GenesisState {
	params := DefaultParams()
	lbi := LastBlockInfo{} // Not setting any value for LastBlockInfo cuz we cant get the last block time without sdk.Context
	return NewGenesisState(params, lbi)
}

// Validate perform object fields validation.
func (m GenesisState) Validate() error {
	if err := m.Params.Validate(); err != nil {
		return err
	}
	if err := m.LastBlockInfo.Validate(); err != nil {
		return err
	}
	return nil
}
