package gov_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/wasmbinding/gov"
	govWbTypes "github.com/archway-network/archway/wasmbinding/gov/types"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// TestGovWASMBindings tests the custom querier for the x/gov WASM bindings.
func TestGovWASMBindings(t *testing.T) {
	// Setup
	chain := e2eTesting.NewTestChain(t, 1)
	ctx, keeper := chain.GetContext(), chain.GetApp().GovKeeper

	// Create custom plugins
	queryPlugin := gov.NewQueryHandler(keeper)

	proposalId := govTypes.DefaultStartingProposalID
	textProposal := govTypes.NewTextProposal("foo", "bar")

	anyTime := time.Now().UTC()
	proposal, pErr := govTypes.NewProposal(textProposal, proposalId, anyTime, anyTime)
	require.NoError(t, pErr)

	accAddrs, _ := e2eTesting.GenAccounts(2)
	depositor := accAddrs[0]
	voter := accAddrs[1]

	t.Run("Query non-existing proposals", func(t *testing.T) {
		query := govWbTypes.ProposalsRequest{}

		res, err := queryPlugin.GetProposals(ctx, query)
		require.NoError(t, err)
		assert.Empty(t, res.Proposals)
	})

	// Store a proposal
	keeper.SetProposal(ctx, proposal)

	t.Run("Query existing proposals", func(t *testing.T) {
		query := govWbTypes.ProposalsRequest{}

		res, err := queryPlugin.GetProposals(ctx, query)
		require.NoError(t, err)
		assert.Len(t, res.Proposals, 1)
	})

	// Make a deposit
	deposit := govTypes.NewDeposit(proposalId, depositor, nil)
	keeper.SetDeposit(ctx, deposit)

	t.Run("Query proposals by depositor", func(t *testing.T) {
		query := govWbTypes.ProposalsRequest{
			Depositor: depositor.String(),
		}

		res, err := queryPlugin.GetProposals(ctx, query)
		require.NoError(t, err)
		assert.Len(t, res.Proposals, 1)
	})

	// Vote
	keeper.ActivateVotingPeriod(ctx, proposal)
	vote := govTypes.NewVote(proposalId, voter, govTypes.NewNonSplitVoteOption(govTypes.OptionYes))
	keeper.SetVote(ctx, vote)

	t.Run("Query proposals by voter", func(t *testing.T) {
		query := govWbTypes.ProposalsRequest{
			Voter: voter.String(),
		}

		res, err := queryPlugin.GetProposals(ctx, query)
		require.NoError(t, err)
		assert.Len(t, res.Proposals, 1)
	})

	t.Run("Query proposals by status", func(t *testing.T) {
		query := govWbTypes.ProposalsRequest{
			Status: govTypes.StatusVotingPeriod.String(),
		}

		res, err := queryPlugin.GetProposals(ctx, query)
		require.NoError(t, err)
		assert.Len(t, res.Proposals, 1)
	})

	t.Run("Query invalid status", func(t *testing.T) {
		query := govWbTypes.ProposalsRequest{
			Status: "INVALID",
		}

		_, err := queryPlugin.GetProposals(ctx, query)
		require.Error(t, err)
	})

	t.Run("Query vote on proposal", func(t *testing.T) {
		query := govWbTypes.VoteRequest{
			ProposalId: proposalId,
			Voter:      voter.String(),
		}

		res, err := queryPlugin.GetVote(ctx, query)
		require.NoError(t, err)
		assert.Equal(t, proposalId, res.Vote.ProposalId)
		assert.Equal(t, voter.String(), res.Vote.Voter)
		assert.NotEmpty(t, res.Vote.Options)
	})

	t.Run("Query vote on invalid proposal", func(t *testing.T) {
		query := govWbTypes.VoteRequest{
			ProposalId: 2,
			Voter:      voter.String(),
		}

		_, err := queryPlugin.GetVote(ctx, query)
		assert.ErrorContains(t, err, "vote not found for proposal")
	})
}
