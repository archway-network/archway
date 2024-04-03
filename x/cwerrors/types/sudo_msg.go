package types

import "encoding/json"

// SudoMsg is message sent to a contract.
// This is encoded as JSON input to the contract when executing the callback
type SudoMsg struct {
	// Error is the endpoint name at the contract which is called
	Error *SudoError `json:"error,omitempty"`
}

// NewSudoMsg creates a new SudoMsg instance.
func NewSudoMsg(sudoErr SudoError) SudoMsg {
	return SudoMsg{
		Error: &sudoErr,
	}
}

// Bytes returns the sudo message as JSON bytes
func (s SudoMsg) Bytes() []byte {
	msgBz, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return msgBz
}

// String returns the sudo message as JSON string
func (s SudoMsg) String() string {
	return string(s.Bytes())
}
