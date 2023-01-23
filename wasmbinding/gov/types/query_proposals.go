package types

import (
	"fmt"
	"time"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// ProposalsRequest is the Query.ProposalRequest request.
type ProposalsRequest struct {
	// Voter is the bech32 encoded account address of the voter.
	Voter string `json:"voter,omitempty"`
	// Depositor is the bech32 encoded account address of the voter.
	Depositor string `json:"depositor,omitempty"`
	// Status is the status from the enum govTypes.ProposalStatus.
	Status string `json:"status,omitempty"`
	// Page is an optional argument to paginate the request.
	Page int `json:"page,omitempty"`
	// Limit is an optional argument to paginate the request.
	Limit int `json:"limit,omitempty"`
}

type (
	ProposalsResponse struct {
		// Proposals is the list of proposals returned by the query.
		Proposals []Proposal `json:"proposals"`
	}

	Proposal struct {
		// ProposalId is the proposal identifier.
		ProposalId uint64 `json:"proposal_id"`
		// Status is the proposal status.
		Status string `json:"status"`
		// FinalTallyResult is the final tally of the vote.
		FinalTallyResult govTypes.TallyResult `json:"final_tally_result"`
		// SubmitTime is the proposal submission time.
		SubmitTime string `json:"submit_time"`
		// DepositEndTime is the proposal deposit end time.
		DepositEndTime string `json:"deposit_end_time"`
		// TotalDeposit is the proposal total deposit.
		TotalDeposit wasmvmtypes.Coins `json:"total_deposit"`
		// VotingStartTime is the proposal voting start time.
		VotingStartTime string `json:"voting_start_time"`
		// VotingEndTime is the proposal voting end time.
		VotingEndTime string `json:"voting_end_time"`
	}
)

// Validate performs request fields validation.
func (r ProposalsRequest) Validate() error {
	if r.Voter != "" {
		if _, err := sdk.AccAddressFromBech32(r.Voter); err != nil {
			return fmt.Errorf("voter: parsing: %w", err)
		}
	}

	if r.Depositor != "" {
		if _, err := sdk.AccAddressFromBech32(r.Depositor); err != nil {
			return fmt.Errorf("depositor: parsing: %w", err)
		}
	}

	if r.Status != "" {
		if _, err := govTypes.ProposalStatusFromString(r.Status); err != nil {
			return fmt.Errorf("status: parsing: %w", err)
		}
	}

	return nil
}

// GetVoter returns the rewards address as sdk.AccAddress or nil.
func (r ProposalsRequest) GetVoter() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(r.Voter)
	if err != nil {
		return nil
	}

	return addr
}

// GetDepositor returns the rewards address as sdk.AccAddress or nil.
func (r ProposalsRequest) GetDepositor() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(r.Depositor)
	if err != nil {
		return nil
	}

	return addr
}

func (r ProposalsRequest) GetPage() int {
	if r.Page == 0 {
		return 1
	}

	return r.Page
}

func NewProposalsResponse(proposals []govTypes.Proposal) ProposalsResponse {
	resp := ProposalsResponse{
		Proposals: make([]Proposal, 0, len(proposals)),
	}

	for _, proposal := range proposals {
		resp.Proposals = append(resp.Proposals, Proposal{
			ProposalId:       proposal.ProposalId,
			Status:           proposal.Status.String(),
			FinalTallyResult: proposal.FinalTallyResult,
			SubmitTime:       proposal.SubmitTime.Format(time.RFC3339Nano),
			DepositEndTime:   proposal.DepositEndTime.Format(time.RFC3339Nano),
			TotalDeposit:     wasmdTypes.NewWasmCoins(proposal.TotalDeposit),
			VotingStartTime:  proposal.VotingStartTime.Format(time.RFC3339Nano),
			VotingEndTime:    proposal.VotingEndTime.Format(time.RFC3339Nano),
		})
	}

	return resp
}
