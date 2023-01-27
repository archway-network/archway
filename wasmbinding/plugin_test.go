package wasmbinding_test

import (
	"fmt"
	"testing"
	"time"

	wasmVmTypes "github.com/CosmWasm/wasmvm/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/wasmbinding"
)

// TestWASMBindingPlugins tests common failure scenarios for custom querier and msg handler plugins.
// Happy paths are tested in the integration tests.
func TestWASMBindingPlugins(t *testing.T) {
	// Setup
	chain := e2eTesting.NewTestChain(t, 1)
	mockMessenger := testutils.NewMockMessenger()
	mockContractAddr := e2eTesting.GenContractAddresses(1)[0]
	ctx := chain.GetContext()

	// Create custom plugins
	rewardsKeeper := chain.GetApp().RewardsKeeper
	govKeeper := chain.GetApp().GovKeeper
	msgPlugin := wasmbinding.BuildWasmMsgDecorator(rewardsKeeper)
	queryPlugin := wasmbinding.BuildWasmQueryPlugin(rewardsKeeper, govKeeper)

	// Querier tests
	t.Run("Querier failure", func(t *testing.T) {
		t.Run("Invalid JSON request", func(t *testing.T) {
			_, err := queryPlugin.Custom(ctx, []byte("invalid"))
			assert.ErrorIs(t, err, sdkErrors.ErrInvalidRequest)
		})

		t.Run("Invalid request (not one of)", func(t *testing.T) {
			queryBz := []byte("{}")

			_, err := queryPlugin.Custom(ctx, queryBz)
			assert.ErrorIs(t, err, sdkErrors.ErrInvalidRequest)
		})
	})

	t.Run("Querier OK", func(t *testing.T) {
		t.Run("Query empty metada", func(t *testing.T) {
			_, err := queryPlugin.Custom(ctx, []byte("{\"contract_metadata\": {\"contract_address\": \""+mockContractAddr.String()+"\"}}"))
			assert.Error(t, err)
		})

		t.Run("Query empty rewards", func(t *testing.T) {
			_, err := queryPlugin.Custom(ctx, []byte("{\"rewards_records\": {\"rewards_address\": \""+mockContractAddr.String()+"\"}}"))
			require.NoError(t, err)
		})

		t.Run("Query gov vote", func(t *testing.T) {
			proposalId := govTypes.DefaultStartingProposalID
			textProposal := govTypes.NewTextProposal("foo", "bar")

			anyTime := time.Now().UTC()
			proposal, pErr := govTypes.NewProposal(textProposal, proposalId, anyTime, anyTime)
			require.NoError(t, pErr)
			govKeeper.SetProposal(ctx, proposal)

			accAddrs, _ := e2eTesting.GenAccounts(2)
			depositor := accAddrs[0]
			deposit := govTypes.NewDeposit(proposalId, depositor, nil)
			govKeeper.SetDeposit(ctx, deposit)

			voter := accAddrs[1]
			govKeeper.ActivateVotingPeriod(ctx, proposal)
			vote := govTypes.NewVote(proposalId, voter, govTypes.NewNonSplitVoteOption(govTypes.OptionYes))
			govKeeper.SetVote(ctx, vote)

			_, err := queryPlugin.Custom(ctx, []byte(fmt.Sprintf("{\"gov_vote\": {\"proposal_id\": %d, \"voter\": \"%s\"}}", proposalId, voter)))
			require.NoError(t, err)
		})
	})

	// Msg handler tests
	t.Run("MsgHandler failure", func(t *testing.T) {
		t.Run("Invalid JSON request", func(t *testing.T) {
			msg := wasmVmTypes.CosmosMsg{
				Custom: []byte("invalid"),
			}
			_, _, err := msgPlugin(mockMessenger).DispatchMsg(ctx, mockContractAddr, "", msg)
			assert.ErrorIs(t, err, sdkErrors.ErrInvalidRequest)
		})

		t.Run("Invalid request (not one of)", func(t *testing.T) {
			msg := wasmVmTypes.CosmosMsg{
				Custom: []byte("{}"),
			}
			_, _, err := msgPlugin(mockMessenger).DispatchMsg(ctx, mockContractAddr, "", msg)
			assert.ErrorIs(t, err, sdkErrors.ErrInvalidRequest)
		})
	})

	t.Run("MsgHandler OK", func(t *testing.T) {
		t.Run("No-op (non-custom msg)", func(t *testing.T) {
			msg := wasmVmTypes.CosmosMsg{}
			_, _, err := msgPlugin(mockMessenger).DispatchMsg(ctx, mockContractAddr, "", msg)
			assert.NoError(t, err)
		})
	})
}
