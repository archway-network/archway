package gov_test

// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"

// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

// 	e2eTesting "github.com/archway-network/archway/e2e/testing"
// 	"github.com/archway-network/archway/wasmbinding/gov"
// 	govWbTypes "github.com/archway-network/archway/wasmbinding/gov/types"
// 	extendedGov "github.com/archway-network/archway/x/gov"
// )

// // TestGovWASMBindings tests the custom querier for the x/gov WASM bindings.
// func TestGovWASMBindings(t *testing.T) {
// 	// Setup
// 	chain := e2eTesting.NewTestChain(t, 1)
// 	ctx, keeper := chain.GetContext(), chain.GetApp().Keepers.GovKeeper

// 	// Create custom plugins
// 	queryPlugin := gov.NewQueryHandler(extendedGov.NewKeeper(keeper))

// 	accAddrs, _ := e2eTesting.GenAccounts(2)
// 	depositor := accAddrs[0]
// 	voter := accAddrs[1]

// 	//govAccount := keeper.GetGovernanceAccount(ctx)
// 	params, err := keeper.Params.Get(ctx)
// 	require.NoError(t, err)

// 	// Store a proposal
// 	proposalId := govTypes.DefaultStartingProposalID

// 	proposal, err := govTypes.NewProposal([]sdk.Msg{}, proposalId, ctx.BlockHeader().Time, ctx.BlockHeader().Time.Add(*params.MaxDepositPeriod), "", "Text Proposal", "Description", depositor, false)
// 	require.NoError(t, err)
// 	err = keeper.SetProposal(ctx, proposal)
// 	require.NoError(t, err)

// 	// Make a deposit
// 	deposit := govTypes.NewDeposit(proposalId, depositor, nil)
// 	err = keeper.SetDeposit(ctx, deposit)
// 	require.NoError(t, err)

// 	// Vote
// 	err = keeper.ActivateVotingPeriod(ctx, proposal)
// 	require.NoError(t, err)
// 	err = keeper.AddVote(ctx, proposalId, voter, govTypes.NewNonSplitVoteOption(govTypes.OptionYes), "")
// 	require.NoError(t, err)

// 	t.Run("Query vote on proposal", func(t *testing.T) {
// 		query := govWbTypes.VoteRequest{
// 			ProposalID: proposalId,
// 			Voter:      voter.String(),
// 		}

// 		res, err := queryPlugin.GetVote(ctx, query)
// 		require.NoError(t, err)
// 		assert.Equal(t, proposalId, res.Vote.ProposalId)
// 		assert.Equal(t, voter.String(), res.Vote.Voter)
// 		assert.NotEmpty(t, res.Vote.Options)
// 	})

// 	t.Run("Query vote on invalid proposal", func(t *testing.T) {
// 		query := govWbTypes.VoteRequest{
// 			ProposalID: 2,
// 			Voter:      voter.String(),
// 		}

// 		_, err := queryPlugin.GetVote(ctx, query)
// 		assert.ErrorContains(t, err, "vote not found for proposal")
// 	})
// }
