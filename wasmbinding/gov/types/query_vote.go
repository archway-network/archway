package types

import (
	"fmt"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type VoteRequest struct {
	// ProposalID is the unique ID of the proposal.
	ProposalId uint64 `json:"proposal_id"`

	// Voter is the bech32 encoded account address of the voter.
	Voter string `json:"voter"`
}

type (
	VoteResponse struct {
		// Vote defines a vote on a governance proposal.
		Vote Vote `json:"vote"`
	}

	Vote struct {
		// ProposalId is the proposal identifier.
		ProposalId uint64 `json:"proposal_id"`
		// Voter is the bech32 encoded account address of the voter.
		Voter string `json:"voter"`
		// Option is the voting option from the enum.
		Options govTypes.WeightedVoteOptions `json:"option"`
	}
)

// Validate performs request fields validation.
func (r VoteRequest) Validate() error {
	if r.ProposalId == 0 {
		return fmt.Errorf("proposal_id: the proposal ID is mandatory")
	}

	if _, err := sdk.AccAddressFromBech32(r.Voter); err != nil {
		return fmt.Errorf("voter: parsing: %w", err)
	}

	return nil
}

func NewVoteResponse(vote govTypes.Vote) VoteResponse {
	resp := VoteResponse{
		Vote: Vote{
			ProposalId: vote.ProposalId,
			Voter:      vote.Voter,
			Options:    vote.Options,
		},
	}

	return resp
}
