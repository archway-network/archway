package types

import (
	"fmt"
	"time"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/archway-network/archway/wasmbinding/pkg"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// ProposalsRequest is the Query.ProposalRequest request.
type ProposalsRequest struct {
	// Voter is the bech32 encoded account address of the voter.
	Voter string `json:"voter"`
	// Depositor is the bech32 encoded account address of the voter.
	Depositor string `json:"depositor"`
	// Status is the status from the enum govTypes.ProposalStatus.
	Status string `json:"status"`
	// Pagination is an optional pagination options for the request.
	// Limit should not exceed the MaxWithdrawRecords param value.
	Pagination *pkg.PageRequest `json:"pagination"`
}

type (
	ProposalsResponse struct {
		// Proposals is the list of proposals returned by the query.
		Proposals []Proposal `json:"proposals"`
		// Pagination is the pagination details in the response.
		Pagination pkg.PageResponse `json:"pagination"`
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

// NewProposalsResponse builds a new ProposalsResponse.
func NewProposalsResponse(proposals []govTypes.Proposal, pageResp query.PageResponse) ProposalsResponse {
	resp := ProposalsResponse{
		Proposals:  make([]Proposal, 0, len(proposals)),
		Pagination: pkg.NewPageResponseFromSDK(pageResp),
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
