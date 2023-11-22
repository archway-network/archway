package types

import "encoding/json"

// SudoMsg callback message sent to a contract.
type SudoMsg struct {
	Callback *CallbackMsg `json:"callback,omitempty"`
}

// CallbackMsg
type CallbackMsg struct {
	JobID uint64 `json:"job_id"`
}

// NewCallback creates a new Callback instance.
func NewCallbackMsg(jobID uint64) SudoMsg {
	return SudoMsg{
		Callback: &CallbackMsg{
			JobID: jobID,
		},
	}
}

func (s SudoMsg) Bytes() []byte {
	msgBz, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return msgBz
}

func (s SudoMsg) String() string {
	return string(s.Bytes())
}
