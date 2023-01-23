package gov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/archway-network/archway/wasmbinding/gov/types"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// KeeperReaderExpected defines the x/gov keeper expected read operations.
type KeeperReaderExpected interface {
	GetProposalsFiltered(c sdk.Context, params govTypes.QueryProposalsParams) govTypes.Proposals
	GetVote(c sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) (vote govTypes.Vote, found bool)
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

	proposalStatus, _ := govTypes.ProposalStatusFromString(req.Status)
	params := govTypes.NewQueryProposalsParams(req.GetPage(), req.Limit, proposalStatus, req.GetVoter(), req.GetDepositor())
	proposals := h.govKeeper.GetProposalsFiltered(ctx, params)

	return types.NewProposalsResponse(proposals), nil
}

// GetVote returns the vote weighted options for a given proposal and voter.
func (h QueryHandler) GetVote(ctx sdk.Context, req types.VoteRequest) (types.VoteResponse, error) {
	if err := req.Validate(); err != nil {
		return types.VoteResponse{}, fmt.Errorf("vote: %w", err)
	}

	vote, found := h.govKeeper.GetVote(ctx, req.ProposalId, req.MustGetVoter())
	if !found {
		err := sdkErrors.Wrap(govTypes.ErrInvalidVote, fmt.Errorf("vote not found for proposal %d and voter %s", req.ProposalId, req.Voter).Error())
		return types.VoteResponse{}, err
	}

	return types.NewVoteResponse(vote), nil
}
