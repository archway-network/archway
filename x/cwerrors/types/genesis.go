package types

// NewGenesisState creates a new GenesisState object.
func NewGenesisState(
	params Params,
) *GenesisState {
	return &GenesisState{
		Params: params,
		Errors: nil,
	}
}

// DefaultGenesisState returns a default genesis state.
func DefaultGenesis() *GenesisState {
	defaultParams := DefaultParams()
	return NewGenesisState(defaultParams)
}

// Validate perform object fields validation.
func (g GenesisState) Validate() error {
	return g.Params.Validate()
}
