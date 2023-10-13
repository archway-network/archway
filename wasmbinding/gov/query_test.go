package gov_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/wasmbinding/gov"
	govWbTypes "github.com/archway-network/archway/wasmbinding/gov/types"
)

// TestGovWASMBindings tests the custom querier for the x/gov WASM bindings.
func TestGovWASMBindings(t *testing.T) {
	// Setup
	chain := e2eTesting.NewTestChain(t, 1)
	ctx, keeper := chain.GetContext(), chain.GetApp().Keepers.GovKeeper

	// Create custom plugins
	queryPlugin := gov.NewQueryHandler(keeper)

	accAddrs, _ := e2eTesting.GenAccounts(2)
	depositor := accAddrs[0]
	voter := accAddrs[1]

	//govAccount := keeper.GetGovernanceAccount(ctx)
	params := keeper.GetParams(ctx)

	// Store a proposal
	proposalId := govTypes.DefaultStartingProposalID

	proposal, err := govTypes.NewProposal([]sdk.Msg{}, proposalId, ctx.BlockHeader().Time, ctx.BlockHeader().Time.Add(*params.MaxDepositPeriod), "", "Text Proposal", "Description", depositor)
	require.NoError(t, err)
	keeper.SetProposal(ctx, proposal)

	// Make a deposit
	deposit := govTypes.NewDeposit(proposalId, depositor, nil)
	keeper.SetDeposit(ctx, deposit)

	// Vote
	keeper.ActivateVotingPeriod(ctx, proposal)
	vote := govTypes.NewVote(proposalId, voter, govTypes.NewNonSplitVoteOption(govTypes.OptionYes), "")
	keeper.SetVote(ctx, vote)

	t.Run("Query vote on proposal", func(t *testing.T) {
		query := govWbTypes.VoteRequest{
			ProposalID: proposalId,
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
			ProposalID: 2,
			Voter:      voter.String(),
		}

		_, err := queryPlugin.GetVote(ctx, query)
		assert.ErrorContains(t, err, "vote not found for proposal")
	})
}
