package gov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	wasmTypes "github.com/archway-network/archway/wasmbinding/gov/types"
)

// KeeperReaderExpected defines the x/gov keeper expected read operations.
type KeeperReaderExpected interface {
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

// GetVote returns the vote weighted options for a given proposal and voter.
func (h QueryHandler) GetVote(ctx sdk.Context, req wasmTypes.VoteRequest) (wasmTypes.VoteResponse, error) {
	if err := req.Validate(); err != nil {
		return wasmTypes.VoteResponse{}, fmt.Errorf("vote: %w", err)
	}

	vote, found := h.govKeeper.GetVote(ctx, req.ProposalID, req.MustGetVoter())
	if !found {
		err := errorsmod.Wrap(types.ErrInvalidVote, fmt.Errorf("vote not found for proposal %d and voter %s", req.ProposalID, req.Voter).Error())
		return wasmTypes.VoteResponse{}, err
	}

	return wasmTypes.NewVoteResponse(vote), nil
}
