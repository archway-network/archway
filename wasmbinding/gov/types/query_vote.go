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
		Options []WeightedVoteOption `json:"option"`
	}

	WeightedVoteOption struct {
		Option string  `json:"option,omitempty"`
		Weight sdk.Dec `json:"weight,omitempty"`
	}
)

// Validate performs request fields validation.
func (r VoteRequest) Validate() error {
	if r.ProposalId == 0 {
		return fmt.Errorf("proposal_id: must specify a proposal ID to query")
	}

	if _, err := sdk.AccAddressFromBech32(r.Voter); err != nil {
		return fmt.Errorf("voter: parsing: %w", err)
	}

	return nil
}

// MustGetVoter returns the voter as sdk.AccAddress.
// CONTRACT: panics in case of an error (should not happen since we validate the request).
func (r VoteRequest) MustGetVoter() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(r.Voter)
	if err != nil {
		// Should not happen since we validate the request before this call
		panic(fmt.Errorf("wasm bindings: voteRequest request: parsing voter: %w", err))
	}

	return addr
}

func NewVoteResponse(vote govTypes.Vote) VoteResponse {
	resp := VoteResponse{
		Vote: Vote{
			ProposalId: vote.ProposalId,
			Voter:      vote.Voter,
			Options:    make([]WeightedVoteOption, 0, len(vote.Options)),
		},
	}

	for _, option := range vote.Options {
		resp.Vote.Options = append(resp.Vote.Options, WeightedVoteOption{
			Option: option.String(),
			Weight: option.Weight,
		})
	}

	return resp
}
