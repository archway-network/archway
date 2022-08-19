package types

// MsgInstantiate is handled by the Instantiate entrypoint.
type MsgInstantiate struct {
	// Params are the contract parameters.
	Params Params
}
