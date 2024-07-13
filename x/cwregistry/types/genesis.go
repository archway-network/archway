package types

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for _, meta := range gs.CodeMetadata {
		if err := meta.Validate(); err != nil {
			return err
		}
	}
	return nil
}
