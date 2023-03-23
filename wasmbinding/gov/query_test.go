package gov_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	e2etesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/wasmbinding/gov"
	govWbTypes "github.com/archway-network/archway/wasmbinding/gov/types"
)

// TestGovWASMBindings tests the custom querier for the x/gov WASM bindings.
func TestGovWASMBindings(t *testing.T) {
	// Setup
	chain := e2etesting.NewTestChain(t, 1)
	ctx, keeper := chain.GetContext(), chain.GetApp().GovKeeper

	// Create custom plugins
	queryPlugin := gov.NewQueryHandler(keeper)

	accAddrs, _ := e2etesting.GenAccounts(2)
	depositor := accAddrs[0]
	voter := accAddrs[1]

	// Store a proposal
	proposalID := govTypes.DefaultStartingProposalID
	textProposal := govTypes.NewTextProposal("foo", "bar")
	proposal, pErr := govTypes.NewProposal(textProposal, proposalID, time.Now().UTC(), time.Now().UTC())
	require.NoError(t, pErr)
	keeper.SetProposal(ctx, proposal)

	// Make a deposit
	deposit := govTypes.NewDeposit(proposalID, depositor, nil)
	keeper.SetDeposit(ctx, deposit)

	// Vote
	keeper.ActivateVotingPeriod(ctx, proposal)
	vote := govTypes.NewVote(proposalID, voter, govTypes.NewNonSplitVoteOption(govTypes.OptionYes))
	keeper.SetVote(ctx, vote)

	t.Run("Query vote on proposal", func(t *testing.T) {
		query := govWbTypes.VoteRequest{
			ProposalID: proposalID,
			Voter:      voter.String(),
		}

		res, err := queryPlugin.GetVote(ctx, query)
		require.NoError(t, err)
		assert.Equal(t, proposalID, res.Vote.ProposalID)
		assert.Equal(t, voter.String(), res.Vote.Voter)
		assert.NotEmpty(t, res.Vote.Options)
	})

	t.Run("Query vote on invalid proposal", func(t *testing.T) {
		query := govWbTypes.VoteRequest{
			ProposalID: 2,
			Voter:      voter.String(),
		}

		_, err := queryPlugin.GetVote(ctx, query)
		assert.ErrorContains(t, err, "vote not found for proposal")
	})
}
