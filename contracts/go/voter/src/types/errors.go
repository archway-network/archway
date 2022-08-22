package types

import (
	"errors"
)

var (
	_ error = ErrInternal{}
	_ error = ErrInvalidRequest{}
	_ error = ErrUnimplemented{}

	// ErrVotingClosed indicates that a specific voting is closed for new votes.
	ErrVotingClosed = errors.New("voting is closed")

	// ErrAlreadyVoted indicates that a sender has already voted.
	ErrAlreadyVoted = errors.New("already voted")
)

// ErrInternal indicates that something went wrong (should not happen).
// tinyjson:skip
type ErrInternal struct {
	Msg string
}

// NewErrInternal ...
func NewErrInternal(msg string) ErrInternal {
	return ErrInternal{msg}
}

// Error implements the error interface.
func (e ErrInternal) Error() string {
	return "internal error: " + e.Msg
}

// ErrInvalidRequest indicates that user input is incorrect.
// tinyjson:skip
type ErrInvalidRequest struct {
	Msg string
}

// NewErrInvalidRequest ...
func NewErrInvalidRequest(msg string) ErrInvalidRequest {
	return ErrInvalidRequest{msg}
}

// Error implements the error interface.
func (e ErrInvalidRequest) Error() string {
	return "invalid request: " + e.Msg
}

// ErrUnimplemented indicates that function call is not implemented yet.
// tinyjson:skip
type ErrUnimplemented struct {
	Method string
}

// NewErrUnimplemented ...
func NewErrUnimplemented(methodName string) ErrUnimplemented {
	return ErrUnimplemented{methodName}
}

// Error implements the error interface.
func (e ErrUnimplemented) Error() string {
	return "method is not implemented: " + e.Method
}
