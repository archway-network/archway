package interchaintest

import (
	"context"
	"testing"

	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/relayer"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestCWICA tests the CWICA functionality.
// It tests the following:
// 1. Registering an interchain account on the counterparty chain via a cw smart contract on the archway chain => Contract receives ica address and stores it
// 2. Trying to register the same interchain account again => Tx fails with error
// 3. Submitting a tx on the contract which will vote on the proposal on counterparty chain - There is no proposal on chain => Contract receives error details
// 4. Submitting a tx on the contract which will vote on the proposal on counterparty chain - There is a proposal on chain => Contract receives vote ack
// 5. Submitting a tx on the contract which will vote on the proposal on counterparty chain - The message timeout is too small => Contract receives timeout msg
// 6. Submitting a tx on the contract which will vote on the proposal on counterparty chain - The channel is closed => Tx fails with error
// 7. Registering the interchain account again to open the channel => Contract receives ica address and stores it. Channel is opened
// 8. Submitting a tx on the contract which will vote on the proposal on counterparty chain - Vote is different => Contract receives vote ack
func TestCWICA(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	numOfVals := 1
	junoChainSpec := &interchaintest.ChainSpec{
		Name:          "juno",
		ChainName:     "juno",
		Version:       "v25.0.0",
		NumValidators: &numOfVals,
		ChainConfig: ibc.ChainConfig{
			GasAdjustment: 2,
		},
	}
	archwayChainSpec := GetArchwaySpec("local", numOfVals)

	// Setup the chains
	chainFactory := interchaintest.NewBuiltinChainFactory(
		zaptest.NewLogger(t),
		[]*interchaintest.ChainSpec{
			archwayChainSpec,
			junoChainSpec,
		})
	chains, err := chainFactory.Chains(t.Name())
	require.NoError(t, err)
	archwayChain, counterpartyChain := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	// Setup the relayer
	const (
		path        = "interchainx-ica-test-path"
		relayerName = "relayer"
	)
	relayerFactory := interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t),
		relayer.StartupFlags("-b", "100", "-p", "events"),
		// relayer.ImagePull(false),
		// relayer.DockerImage(&ibc.DockerImage{
		// 	Repository: "ghcr.io/cosmos/relayer",
		// 	Version:    "justin-CoC", // sha256:a7f03cc955c1bd8d1436bee29eaf5c1e44298e17d1dfb3fecb1be912f206819b
		// }),
	)
	client, network := interchaintest.DockerSetup(t)
	relayer := relayerFactory.Build(t, client, network)

	// Create the IBC network with the chains and relayer
	ibcChannelOpts := ibc.DefaultChannelOpts()
	ibcChannelOpts.Order = ibc.Ordered
	ic := interchaintest.NewInterchain().
		AddChain(archwayChain).
		AddChain(counterpartyChain).
		AddRelayer(relayer, relayerName).
		AddLink(interchaintest.InterchainLink{
			Chain1:            archwayChain,
			Chain2:            counterpartyChain,
			Relayer:           relayer,
			Path:              path,
			CreateChannelOpts: ibcChannelOpts,
		})
	ir := cosmos.DefaultEncoding().InterfaceRegistry
	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)
	ctx := context.Background()

	// Starts all the components of the IBC network
	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false,
	}))

	archwayChainUser := fundChainUser(t, ctx, archwayChain)
	counterpartyChainUser := fundChainUser(t, ctx, counterpartyChain)

	connections, err := relayer.GetConnections(ctx, eRep, archwayChain.Config().ChainID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(connections), 1)
	connection := connections[0]
	err = relayer.StartRelayer(ctx, eRep, path)
	require.NoError(t, err)

	t.Cleanup(func() {
		err = ic.Close()
		if err != nil {
			t.Logf("an error occurred while closing the interchain: %s", err)
		}
	})

	// Upload the contract to archway chain
	codeId, err := archwayChain.StoreContract(ctx, archwayChainUser.KeyName(), "artifacts/cwica.wasm")
	require.NoError(t, err)
	require.NotEmpty(t, codeId)

	// Instantiate the contract
	initMsg := "{\"connection_id\":\"" + connection.ID + "\"}"
	contractAddress, err := InstantiateContract(archwayChain, archwayChainUser, ctx, codeId, initMsg)
	require.NoError(t, err)

	// Dump state of the contract and ensure the contract is in the expected state
	var contractRes cwicaContractResponse
	err = archwayChain.QueryContract(ctx, contractAddress, QueryMsg{DumpState: &struct{}{}}, &contractRes)
	require.NoError(t, err)
	require.Equal(t, archwayChainUser.FormattedAddress(), contractRes.Data.Owner)
	require.Equal(t, connection.ID, contractRes.Data.ConnectionId)

	// Register a new interchain account on the counterparty chain
	execMsg := `{"register":{}}`
	err = ExecuteContract(archwayChain, archwayChainUser, ctx, contractAddress, execMsg)
	require.NoError(t, err)

	aH, err := archwayChain.Height(ctx)
	require.NoError(t, err)

	// Wait for the MsgChannelOpenAck on archway chain
	_, err = cosmos.PollForMessage(ctx, archwayChain, ir, aH, aH+10, func(found *channeltypes.MsgChannelOpenAck) bool {
		t.Log(found)
		return found.PortId == "icacontroller-"+contractAddress
	})
	require.NoError(t, err)

	// Get the address of the ica account address of the counterparty chain which has just been registered
	icaCounterpartyAddress, err := GetInterchainAccountAddress(archwayChain, ctx, contractAddress, connection.ID)
	require.NoError(t, err)

	// Ensure the contract is in the expected state - the ica address should be stored by the contract
	err = archwayChain.QueryContract(ctx, contractAddress, QueryMsg{DumpState: &struct{}{}}, &contractRes)
	require.NoError(t, err)
	require.Equal(t, icaCounterpartyAddress, contractRes.Data.ICAAddress)

	// Ensure an IBC channel is opened between the two chains
	channels, err := relayer.GetChannels(ctx, eRep, archwayChain.Config().ChainID)
	require.NoError(t, err)
	for _, channel := range channels {
		if channel.Counterparty.PortID == "icahost" && channel.PortID == "icacontroller-"+contractAddress {
			require.Equal(t, "STATE_OPEN", channel.State)
		}
	}

	// Trying to register the same interchain account again should error out as channel already exists
	err = ExecuteContract(archwayChain, archwayChainUser, ctx, contractAddress, execMsg)
	require.Error(t, err)

	// Register the contract for errors
	err = RegisterContractForError(archwayChain, archwayChainUser, ctx, contractAddress)
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 1, archwayChain)
	require.NoError(t, err)

	// SubmitTx on contract which will vote on the proposal on counterparty chain - There is no proposal on chain. Should error out
	execMsg = `{"vote":{"proposal_id":2,"option":1,"tiny_timeout": false}}`
	err = ExecuteContract(archwayChain, archwayChainUser, ctx, contractAddress, execMsg)
	require.NoError(t, err)

	aH, err = archwayChain.Height(ctx)
	require.NoError(t, err)

	// Wait for the MsgAcknowledgement on the archway chain
	_, err = cosmos.PollForMessage(ctx, archwayChain, ir, aH, aH+10, func(found *channeltypes.MsgAcknowledgement) bool {
		t.Log(found)
		return found.Packet.DestinationPort == "icahost" && found.Packet.SourcePort == "icacontroller-"+contractAddress
	})
	require.NoError(t, err)

	// Ensure the contract is in the expected state - The error on the ica tx should be stored by the contract
	err = archwayChain.QueryContract(ctx, contractAddress, QueryMsg{DumpState: &struct{}{}}, &contractRes)
	require.NoError(t, err)
	require.Contains(t, contractRes.Data.Errors, "error handling packet")

	// Create a gov prop on the counterparty chain
	propMsg, err := counterpartyChain.BuildProposal(
		[]cosmos.ProtoMessage{},
		"TextProp",
		"Summary",
		"Metadata",
		"10000000000"+counterpartyChain.Config().Denom,
		counterpartyChainUser.KeyName(),
		false,
	)
	require.NoError(t, err)
	textProp, err := counterpartyChain.SubmitProposal(ctx, counterpartyChainUser.KeyName(), propMsg)
	require.NoError(t, err)

	// SubmitTx on contract which will vote "YES" on the proposal
	execMsg = `{"vote":{"proposal_id":` + textProp.ProposalID + `,"option":1,"tiny_timeout": false}}`
	err = ExecuteContract(archwayChain, archwayChainUser, ctx, contractAddress, execMsg)
	require.NoError(t, err)

	aH, err = archwayChain.Height(ctx)
	require.NoError(t, err)

	// Wait for the MsgAcknowledgement on the archway chain
	_, err = cosmos.PollForMessage(ctx, archwayChain, ir, aH, aH+10, func(found *channeltypes.MsgAcknowledgement) bool {
		t.Log(found)
		return found.Packet.DestinationPort == "icahost" && found.Packet.SourcePort == "icacontroller-"+contractAddress
	})
	require.NoError(t, err)

	// Fetch the ica user's vote on the counterparty chain. Should be yes.
	vote, err := GetUserVote(counterpartyChain, ctx, textProp.ProposalID, icaCounterpartyAddress)
	require.NoError(t, err)
	require.Equal(t, "VOTE_OPTION_YES", vote.Options[0].Option)

	// Ensure the contract is in the expected state - voted status is true
	err = archwayChain.QueryContract(ctx, contractAddress, QueryMsg{DumpState: &struct{}{}}, &contractRes)
	require.NoError(t, err)
	require.True(t, contractRes.Data.Voted)

	// SubmitTx on contract which will vote "NO" on the proposal -- Very small timeout (1s) so should fail
	execMsg = `{"vote":{"proposal_id":` + textProp.ProposalID + `,"option":3,"tiny_timeout": true}}`
	err = ExecuteContract(archwayChain, archwayChainUser, ctx, contractAddress, execMsg)
	require.NoError(t, err)

	aH, err = archwayChain.Height(ctx)
	require.NoError(t, err)

	// Wait for the MsgTimeout on the archway chain
	_, err = cosmos.PollForMessage(ctx, archwayChain, ir, aH, aH+10, func(found *channeltypes.MsgTimeout) bool {
		t.Log(found)
		return found.Packet.DestinationPort == "icahost" && found.Packet.SourcePort == "icacontroller-"+contractAddress
	})
	require.NoError(t, err)

	// Fetch the ica user's vote on the counterparty chain - Should still be YES as the "no" vote timed out
	vote, err = GetUserVote(counterpartyChain, ctx, textProp.ProposalID, icaCounterpartyAddress)
	require.NoError(t, err)
	require.Equal(t, "VOTE_OPTION_YES", vote.Options[0].Option)

	// Ensure the contract is in the expected state - the timeout state is true
	err = archwayChain.QueryContract(ctx, contractAddress, QueryMsg{DumpState: &struct{}{}}, &contractRes)
	require.NoError(t, err)
	require.True(t, contractRes.Data.Timeout)

	// Get the channels and check if the channel is closed
	channels, err = relayer.GetChannels(ctx, eRep, archwayChain.Config().ChainID)
	require.NoError(t, err)
	for _, channel := range channels {
		if channel.Counterparty.PortID == "icahost" {
			// TODO: ORDER_UNORDERED channels are not closed on msg timeout
			require.Equal(t, "STATE_CLOSED", channel.State)
		}
	}

	// Now with MsgTimeout, the channel is closed. So trying to vote again should error out
	execMsg = `{"vote":{"proposal_id":` + textProp.ProposalID + `,"option":3,"tiny_timeout": false}}`
	err = ExecuteContract(archwayChain, archwayChainUser, ctx, contractAddress, execMsg)
	require.ErrorContains(t, err, "no active channel for this owner")

	// We register the account again to open the channel
	execMsg = `{"register":{}}`
	err = ExecuteContract(archwayChain, archwayChainUser, ctx, contractAddress, execMsg)
	require.NoError(t, err)

	aH, err = archwayChain.Height(ctx)
	require.NoError(t, err)

	// Wait for the MsgChannelOpenAck on archway chain
	_, err = cosmos.PollForMessage(ctx, archwayChain, ir, aH, aH+20, func(found *channeltypes.MsgChannelOpenAck) bool {
		t.Log(found)
		return found.PortId == "icacontroller-"+contractAddress
	})
	require.NoError(t, err)

	// Ensure the contract is in the expected state - the same ica address should be stored by the contract
	err = archwayChain.QueryContract(ctx, contractAddress, QueryMsg{DumpState: &struct{}{}}, &contractRes)
	require.NoError(t, err)
	require.Equal(t, icaCounterpartyAddress, contractRes.Data.ICAAddress)

	// Attempt to vote No on the proposal. Previously it was Yes, now this ica tx should pass
	execMsg = `{"vote":{"proposal_id":` + textProp.ProposalID + `,"option":3,"tiny_timeout": false}}`
	err = ExecuteContract(archwayChain, archwayChainUser, ctx, contractAddress, execMsg)
	require.NoError(t, err)

	aH, err = archwayChain.Height(ctx)
	require.NoError(t, err)

	// Wait for the MsgAcknowledgement on the archway chain
	_, err = cosmos.PollForMessage(ctx, archwayChain, ir, aH, aH+10, func(found *channeltypes.MsgAcknowledgement) bool {
		t.Log(found)
		return found.Packet.DestinationPort == "icahost" && found.Packet.SourcePort == "icacontroller-"+contractAddress
	})
	require.NoError(t, err)

	// Fetch the ica user's vote on the counterparty chain. Should be NO.
	vote, err = GetUserVote(counterpartyChain, ctx, textProp.ProposalID, icaCounterpartyAddress)
	require.NoError(t, err)
	require.Equal(t, "VOTE_OPTION_NO", vote.Options[0].Option)

}

type cwicaContractResponse struct {
	Data cwicaContractResponseObj `json:"data"`
}

type cwicaContractResponseObj struct {
	Owner        string `json:"owner"`
	ConnectionId string `json:"connection_id"`
	ICAAddress   string `json:"ica_address"`
	Voted        bool   `json:"voted"`
	Errors       string `json:"errors"`
	Timeout      bool   `json:"timeout"`
}
