package gov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/archway-network/archway/wasmbinding/gov/types"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// KeeperReaderExpected defines the x/gov keeper expected read operations.
type KeeperReaderExpected interface {
	Proposals(c sdk.Context, req *govTypes.QueryProposalsRequest) (*govTypes.QueryProposalsResponse, error)
	Vote(c sdk.Context, req *govTypes.QueryVoteRequest) (*govTypes.QueryVoteResponse, error)
}

// QueryHandler provides a custom WASM query handler for the x/gov module.
type QueryHandler struct {
	govKeeper KeeperReaderExpected
}

// NewQueryHandler creates a new QueryHandler instance.
func NewQueryHandler(gk KeeperReaderExpected) QueryHandler {
	return QueryHandler{
		govKeeper: gk,
	}
}

// GetProposals returns the paginated list of types.Proposal objects for a given request.
func (h QueryHandler) GetProposals(ctx sdk.Context, req types.ProposalsRequest) (types.ProposalsResponse, error) {
	if err := req.Validate(); err != nil {
		return types.ProposalsResponse{}, fmt.Errorf("proposals: %w", err)
	}

	var pageReq *query.PageRequest
	if req.Pagination != nil {
		req := req.Pagination.ToSDK()
		pageReq = &req
	}

	proposalStatus, _ := govTypes.ProposalStatusFromString(req.Status)
	proposalsReq := govTypes.QueryProposalsRequest{
		Voter:          req.Voter,
		Depositor:      req.Depositor,
		ProposalStatus: proposalStatus,
		Pagination:     pageReq,
	}
	res, err := h.govKeeper.Proposals(ctx, &proposalsReq)
	if err != nil {
		return types.ProposalsResponse{}, err
	}

	return types.NewProposalsResponse(res.Proposals, *res.Pagination), nil
}

// GetVote returns the vote weighted options for a given proposal and voter.
func (h QueryHandler) GetVote(ctx sdk.Context, req types.VoteRequest) (types.VoteResponse, error) {
	if err := req.Validate(); err != nil {
		return types.VoteResponse{}, fmt.Errorf("vote: %w", err)
	}

	voteReq := govTypes.QueryVoteRequest{
		ProposalId: req.ProposalId,
		Voter:      req.Voter,
	}
	res, err := h.govKeeper.Vote(ctx, &voteReq)
	if err != nil {
		return types.VoteResponse{}, err
	}

	return types.NewVoteResponse(res.Vote), nil
}
