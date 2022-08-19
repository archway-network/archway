package types

import "errors"

// MsgIBC is handled by the IBCPacketReceive entrypoint.
type MsgIBC struct {
	// Vote append a new vote to an existing voting (free of charge since we can't transfer funds alongside IBC packet data).
	Vote *IBCVoteRequest `json:",omitempty"`
}

// IBCVoteRequest defines MsgIBC.Vote request.
type IBCVoteRequest struct {
	VoteRequest

	// From is a voter address from another chain.
	From string
}

// Validate performs object fields validation.
func (r IBCVoteRequest) Validate() error {
	if err := r.VoteRequest.Validate(); err != nil {
		return err
	}

	if r.From == "" {
		return errors.New("from: empty")
	}

	return nil
}
