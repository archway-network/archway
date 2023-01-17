package e2eTesting

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
)

// ExecuteGovProposal submits a new proposal and votes for it.
func (chain *TestChain) ExecuteGovProposal(proposerAcc Account, expPass bool, proposalContent govTypes.Content) {
	t := chain.t

	require.NotNil(t, proposalContent)

	// Get params
	k := chain.app.GovKeeper
	depositCoin := k.GetDepositParams(chain.GetContext()).MinDeposit
	votingDur := k.GetVotingParams(chain.GetContext()).VotingPeriod

	// Submit proposal with min deposit to start the voting
	msg, err := govTypes.NewMsgSubmitProposal(proposalContent, depositCoin, proposerAcc.Address)
	require.NoError(t, err)

	_, res, _, _ := chain.SendMsgs(proposerAcc, true, []sdk.Msg{msg})
	txRes := chain.ParseSDKResultData(res)
	require.Len(t, txRes.Data, 1)

	var resp govTypes.MsgSubmitProposalResponse
	require.NoError(t, resp.Unmarshal(txRes.Data[0].Data))
	proposalID := resp.ProposalId

	// Vote with all validators (delegators)
	for i := 0; i < len(chain.valSet.Validators); i++ {
		delegatorAcc := chain.GetAccount(i)

		msg := govTypes.NewMsgVote(delegatorAcc.Address, proposalID, govTypes.OptionYes)
		_, _, _, err = chain.SendMsgs(proposerAcc, true, []sdk.Msg{msg})
		require.NoError(t, err)
	}

	// Wait for voting to end
	chain.NextBlock(votingDur)
	chain.NextBlock(0) // for the Gov EndBlocker to work

	// Check if proposal was passed
	proposal, ok := k.GetProposal(chain.GetContext(), proposalID)
	require.True(t, ok)

	if expPass {
		require.Equal(t, govTypes.StatusPassed.String(), proposal.Status.String())
	} else {
		require.NotEqual(t, govTypes.StatusPassed.String(), proposal.Status.String())
	}
}
