package interchaintest

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"testing"

	cosmosproto "github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/relayer"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestInterchainTxs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	numOfVals := 1
	gaiaChainSpec := &interchaintest.ChainSpec{
		Name:          "juno",
		ChainName:     "juno",
		Version:       "v20.0.0",
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
			gaiaChainSpec,
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
	)
	client, network := interchaintest.DockerSetup(t)
	relayer := relayerFactory.Build(t, client, network)

	// Create the IBC network with the chains and relayer
	ic := interchaintest.NewInterchain().
		AddChain(archwayChain).
		AddChain(counterpartyChain).
		AddRelayer(relayer, relayerName).
		AddLink(interchaintest.InterchainLink{
			Chain1:  archwayChain,
			Chain2:  counterpartyChain,
			Relayer: relayer,
			Path:    path,
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
		SkipPathCreation: true,
	}))
	err = testutil.WaitForBlocks(ctx, 1, archwayChain, counterpartyChain)
	require.NoError(t, err)

	archwayChainUser := fundChainUser(t, ctx, archwayChain)
	counterpartyChainUser := fundChainUser(t, ctx, counterpartyChain)

	// Setting up ibc connections between the two chains
	err = relayer.GeneratePath(ctx, eRep, archwayChain.Config().ChainID, counterpartyChain.Config().ChainID, path)
	require.NoError(t, err)
	err = relayer.CreateClients(ctx, eRep, path, ibc.CreateClientOptions{TrustingPeriod: "336h"})
	require.NoError(t, err)
	err = testutil.WaitForBlocks(ctx, 1, archwayChain, counterpartyChain)
	require.NoError(t, err)
	err = relayer.CreateConnections(ctx, eRep, path)
	require.NoError(t, err)
	err = testutil.WaitForBlocks(ctx, 1, archwayChain, counterpartyChain)
	require.NoError(t, err)
	connections, err := relayer.GetConnections(ctx, eRep, archwayChain.Config().ChainID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(connections), 1)
	connection := connections[0]
	err = relayer.StartRelayer(ctx, eRep, path)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := relayer.StopRelayer(ctx, eRep)
		if err != nil {
			t.Logf("an error occurred while stopping the relayer: %s", err)
		}
		err = ic.Close()
		if err != nil {
			t.Logf("an error occurred while closing the interchain: %s", err)
		}
	})

	// Upload the contract to archway chain
	codeId, err := archwayChain.StoreContract(ctx, archwayChainUser.KeyName(), "artifacts/interchaintxs.wasm")
	require.NoError(t, err)
	require.NotEmpty(t, codeId)

	// Instantiate the contract
	initCounter := 1
	initMsg := "{\"count\":" + strconv.Itoa(initCounter) + ",\"connection_id\":\"" + connection.ID + "\"}"
	contractAddress, err := InstantiateContract(archwayChain, archwayChainUser, ctx, codeId, initMsg)
	require.NoError(t, err)

	// Dump state of the contract
	var contractRes interchaintxsContractResponse
	err = archwayChain.QueryContract(ctx, contractAddress, QueryMsg{DumpState: &struct{}{}}, &contractRes)
	require.NoError(t, err)
	require.NotNil(t, contractRes.Data)
	// Ensure the contract is in the expected state
	require.Equal(t, int32(initCounter), contractRes.Data.Count)
	require.Equal(t, archwayChainUser.FormattedAddress(), contractRes.Data.Owner)
	require.Equal(t, connection.ID, contractRes.Data.ConnectionId)
	require.Equal(t, "", contractRes.Data.CounterpartyVersion)

	// Register a new interchain account on the counterparty chain
	execMsg := `{"register":{}}`
	err = ExecuteContract(archwayChain, archwayChainUser, ctx, contractAddress, execMsg)
	require.NoError(t, err)

	cH, err := counterpartyChain.Height(ctx)
	require.NoError(t, err)
	aH, err := archwayChain.Height(ctx)
	require.NoError(t, err)

	// Wait for the MsgChannelOpenConfirm on the counterparty chain
	_, err = cosmos.PollForMessage(ctx, counterpartyChain, ir, cH, cH+10, func(found *channeltypes.MsgChannelOpenConfirm) bool {
		return found.PortId == "icahost"
	})
	require.NoError(t, err)

	// Wait for the MsgChannelOpenAck on archway chain
	_, err = cosmos.PollForMessage(ctx, archwayChain, ir, aH, aH+10, func(found *channeltypes.MsgChannelOpenAck) bool {
		return found.PortId == "icacontroller-"+contractAddress+"."+strconv.Itoa(initCounter)
	})
	require.NoError(t, err)

	// Get the address of the ica account address of the counterparty chain
	icaCounterpartyAddress, err := GetInterchainAccountAddress(archwayChain, ctx, contractAddress, connection.ID, strconv.Itoa(initCounter))
	require.NoError(t, err)

	err = archwayChain.QueryContract(ctx, contractAddress, QueryMsg{DumpState: &struct{}{}}, &contractRes)
	require.NoError(t, err)
	require.NotNil(t, contractRes.Data)
	// Ensure the contract is in the expected state - the ica address should be stored by the contract
	require.Equal(t, int32(initCounter), contractRes.Data.Count)
	require.Equal(t, archwayChainUser.FormattedAddress(), contractRes.Data.Owner)
	require.Equal(t, connection.ID, contractRes.Data.ConnectionId)
	require.Equal(t, icaCounterpartyAddress, contractRes.Data.CounterpartyVersion)

	// Create a dummy gov prop on the counterparty chain
	propMsg, err := counterpartyChain.BuildProposal([]cosmosproto.Message{}, "TextProp", "Summary", "Metadata", "10000000000"+counterpartyChain.Config().Denom)
	require.NoError(t, err)
	textProp, err := counterpartyChain.SubmitProposal(ctx, counterpartyChainUser.KeyName(), propMsg)
	require.NoError(t, err)

	// SubmitTx on contract which will vote on the proposal
	execMsg = `{"vote":{"proposal_id":` + textProp.ProposalID + `,"option":1}}`
	err = ExecuteContract(archwayChain, archwayChainUser, ctx, contractAddress, execMsg)
	require.NoError(t, err)

	aH, err = archwayChain.Height(ctx)
	require.NoError(t, err)
	cH, err = counterpartyChain.Height(ctx)
	require.NoError(t, err)

	// Wait for the MsgRecvPacket on the counterparty chain
	recv, err := cosmos.PollForMessage(ctx, counterpartyChain, ir, cH, cH+10, func(found *channeltypes.MsgRecvPacket) bool {
		t.Log(found.Packet.SourcePort)
		t.Log(found.Packet.DestinationPort)
		return found.Packet.Sequence == 1
	})
	require.NoError(t, err)
	t.Log(string(recv.Packet.Data))

	// Wait for the MsgAcknowledgement on the archway chain
	ack, err := cosmos.PollForMessage(ctx, archwayChain, ir, aH, aH+10, func(found *channeltypes.MsgAcknowledgement) bool {
		return found.Packet.Sequence == 1 &&
			found.Packet.DestinationPort == "icahost"
	})
	require.NoError(t, err)
	t.Log(string(ack.Acknowledgement))

	// Wait for a while to ensure the relayer picks up the packet
	err = testutil.WaitForBlocks(ctx, 10, archwayChain, counterpartyChain)
	require.NoError(t, err)

	cmd := []string{
		counterpartyChain.Config().Bin, "q", "gov", "vote", textProp.ProposalID, icaCounterpartyAddress,
		"--node", counterpartyChain.GetRPCAddress(),
		"--home", counterpartyChain.HomeDir(),
		"--chain-id", counterpartyChain.Config().ChainID,
		"--output", "json",
	}
	stdout, _, err := counterpartyChain.Exec(ctx, cmd, nil)
	require.NoError(t, err, "could not query the vote")
	require.NotEmpty(t, stdout)
	var propResponse QueryVoteResponse
	err = json.Unmarshal(stdout, &propResponse)
	require.NoError(t, err)
	require.Equal(t, "VOTE_OPTION_YES", propResponse.Options[0].Option)

	err = archwayChain.QueryContract(ctx, contractAddress, QueryMsg{DumpState: &struct{}{}}, &contractRes)
	require.NoError(t, err)
	require.NotNil(t, contractRes.Data)
	// Ensure the contract is in the expected state
	t.Log(contractRes.Data.Voted)

	// h, err := archwayChain.Height(ctx)
	// require.NoError(t, err)
	// err = archwayChain.StopAllNodes(ctx)
	// require.NoError(t, err)
	// state, err := archwayChain.ExportState(ctx, int64(h))
	// require.NoError(t, err)
	// err = ioutil.WriteFile("./testdata/archway_state.json", []byte(state), 0644)
	// require.NoError(t, err)

	h, err := counterpartyChain.Height(ctx)
	require.NoError(t, err)
	err = counterpartyChain.StopAllNodes(ctx)
	require.NoError(t, err)
	state, err := counterpartyChain.ExportState(ctx, int64(h))
	require.NoError(t, err)
	err = ioutil.WriteFile("./testdata/juno_state.json", []byte(state), 0644)
	require.NoError(t, err)
}

type interchaintxsContractResponse struct {
	Data interchaintxsContractResponseObj `json:"data"`
}

type interchaintxsContractResponseObj struct {
	Count               int32  `json:"count"`
	Owner               string `json:"owner"`
	ConnectionId        string `json:"connection_id"`
	CounterpartyVersion string `json:"counterparty_version"`
	Voted               bool   `json:"voted"`
}
