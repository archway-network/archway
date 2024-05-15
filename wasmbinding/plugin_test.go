package wasmbinding_test

import (
	"fmt"
	"testing"
	"time"

	wasmVmTypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/wasmbinding"
	extendedGov "github.com/archway-network/archway/x/gov"
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
	keepers := chain.GetApp().Keepers
	rewardsKeeper := keepers.RewardsKeeper
	govKeeper := keepers.GovKeeper
	msgPlugin := wasmbinding.BuildWasmMsgDecorator(rewardsKeeper)
	queryPlugin := wasmbinding.BuildWasmQueryPlugin(rewardsKeeper, extendedGov.NewKeeper(govKeeper))

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
			accAddrs, _ := e2eTesting.GenAccounts(2)
			depositor := accAddrs[0]

			anyTime := time.Now().UTC()
			proposal, pErr := govTypes.NewProposal([]sdk.Msg{}, proposalId, anyTime, anyTime, "", "Text Proposal", "Description", depositor, false)
			require.NoError(t, pErr)
			govKeeper.SetProposal(ctx, proposal)

			deposit := govTypes.NewDeposit(proposalId, depositor, nil)
			govKeeper.SetDeposit(ctx, deposit)

			voter := accAddrs[1]
			govKeeper.ActivateVotingPeriod(ctx, proposal)
			err := govKeeper.AddVote(ctx, proposalId, voter, govTypes.NewNonSplitVoteOption(govTypes.OptionYes), "")
			require.NoError(t, err)

			_, err = queryPlugin.Custom(ctx, []byte(fmt.Sprintf("{\"gov_vote\": {\"proposal_id\": %d, \"voter\": \"%s\"}}", proposalId, voter)))
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
