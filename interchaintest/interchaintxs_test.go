package interchaintest

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
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
	"gopkg.in/yaml.v2"
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
	cmd := []string{
		"archwayd", "tx", "wasm", "instantiate", codeId, initMsg,
		"--label", "interchaintxs", "--admin", archwayChainUser.FormattedAddress(),
		"--from", archwayChainUser.KeyName(), "--keyring-backend", keyring.BackendTest,
		"--gas", "auto", "--gas-prices", "0aarch", "--gas-adjustment", "2",
		"--node", archwayChain.GetRPCAddress(),
		"--home", archwayChain.HomeDir(),
		"--chain-id", archwayChain.Config().ChainID,
		"--output", "json",
		"-y",
	}
	stdout, _, err := archwayChain.Exec(ctx, cmd, nil)
	require.NoError(t, err)
	require.NotEmpty(t, stdout)

	err = testutil.WaitForBlocks(ctx, 1, archwayChain)
	require.NoError(t, err)

	// Getting the contract address
	cmd = []string{
		"archwayd", "q", "wasm", "list-contract-by-code", codeId,
		"--node", archwayChain.GetRPCAddress(),
		"--home", archwayChain.HomeDir(),
		"--chain-id", archwayChain.Config().ChainID,
	}
	stdout, _, err = archwayChain.Exec(ctx, cmd, nil)
	require.NoError(t, err, "could not list the contracts")
	contactsRes := cosmos.QueryContractResponse{}
	err = yaml.Unmarshal(stdout, &contactsRes)
	require.NoError(t, err, "could not unmarshal query contract response")
	contractAddress := contactsRes.Contracts[0]

	// Dump state of the contract
	var contractRes ContractResponse
	err = archwayChain.QueryContract(ctx, contractAddress, QueryMsg{DumpState: &struct{}{}}, &contractRes)
	require.NoError(t, err)
	require.NotNil(t, contractRes.Data)
	// Ensure the contract is in the expected state
	require.Equal(t, int32(initCounter), contractRes.Data.Count)
	require.Equal(t, archwayChainUser.FormattedAddress(), contractRes.Data.Owner)
	require.Equal(t, connection.ID, contractRes.Data.ConnectionId)
	require.Equal(t, "", contractRes.Data.CounterpartyVersion)

	execMsg := `{"register":{}}`
	cmd = []string{
		"archwayd", "tx", "wasm", "execute", contractAddress, execMsg,
		"--from", archwayChainUser.KeyName(), "--keyring-backend", keyring.BackendTest,
		"--gas", "auto", "--gas-prices", "0aarch", "--gas-adjustment", "2",
		"--node", archwayChain.GetRPCAddress(),
		"--home", archwayChain.HomeDir(),
		"--chain-id", archwayChain.Config().ChainID,
		"--output", "json",
		"-y",
	}
	stdout, _, err = archwayChain.Exec(ctx, cmd, nil)
	require.NoError(t, err)
	require.NotEmpty(t, stdout)

	// Wait for a few blocks on both chains so relayer picks up the packet
	err = testutil.WaitForBlocks(ctx, 10, archwayChain, counterpartyChain)
	require.NoError(t, err)

	// Check if ica has been registered on the archway chain
	ownerAddress := contractAddress
	interchainAccountId := initCounter
	cmd = []string{
		"archwayd", "q", "interchaintxs", "interchain-account", ownerAddress, connection.ID, strconv.Itoa(interchainAccountId),
		"--node", archwayChain.GetRPCAddress(),
		"--home", archwayChain.HomeDir(),
		"--chain-id", archwayChain.Config().ChainID,
		"--output", "json",
	}
	stdout, _, err = archwayChain.Exec(ctx, cmd, nil)
	require.NoError(t, err, "could not query the interchain account")
	require.NotEmpty(t, stdout)

	queryRes := InterchainAccountAccountQueryResponse{}
	err = json.Unmarshal(stdout, &queryRes)
	require.NoError(t, err, "could not parse the interchain account query response")
	require.NotEmpty(t, queryRes.InterchainAccountAddress)
	icaCounterpartyAddress := queryRes.InterchainAccountAddress

	err = archwayChain.QueryContract(ctx, contractAddress, QueryMsg{DumpState: &struct{}{}}, &contractRes)
	require.NoError(t, err)
	require.NotNil(t, contractRes.Data)
	// Ensure the contract is in the expected state
	require.Equal(t, int32(initCounter), contractRes.Data.Count)
	require.Equal(t, archwayChainUser.FormattedAddress(), contractRes.Data.Owner)
	require.Equal(t, connection.ID, contractRes.Data.ConnectionId)
	require.Equal(t, icaCounterpartyAddress, contractRes.Data.CounterpartyVersion)

	// Check the balance of the ica on the counterparty chain
	balance, err := counterpartyChain.GetBalance(ctx, icaCounterpartyAddress, counterpartyChain.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, int64(0), balance.Int64())
	// Send some money to the ica account from faucet
	err = counterpartyChain.SendFunds(ctx, counterpartyChainUser.KeyName(), ibc.WalletAmount{
		Address: icaCounterpartyAddress,
		Denom:   counterpartyChain.Config().Denom,
		Amount:  math.NewInt(1000),
	})
	require.NoError(t, err)
	// Ensure ica account has the funds just sent
	balance, err = counterpartyChain.GetBalance(ctx, icaCounterpartyAddress, counterpartyChain.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, int64(1000), balance.Int64())

	// Create a dummy gov prop on the counterparty chain
	propMsg, err := counterpartyChain.BuildProposal([]cosmosproto.Message{}, "Text Proposal", "Doesnt do anything", "metadata", "10000000000"+counterpartyChain.Config().Denom)
	require.NoError(t, err)
	textProp, err := counterpartyChain.SubmitProposal(ctx, counterpartyChainUser.KeyName(), propMsg)
	require.NoError(t, err)
	testutil.WaitForBlocks(ctx, 1, counterpartyChain)

	// Vote on the proposal as counterpartyChainUser
	cmd = []string{
		counterpartyChain.Config().Bin, "tx", "gov", "vote", textProp.ProposalID, "yes",
		"--from", counterpartyChainUser.KeyName(), "--keyring-backend", keyring.BackendTest,
		"--gas", "auto", "--gas-prices", "1" + counterpartyChain.Config().Denom, "--gas-adjustment", "2",
		"--node", counterpartyChain.GetRPCAddress(),
		"--home", counterpartyChain.HomeDir(),
		"--chain-id", counterpartyChain.Config().ChainID,
		"--output", "json",
		"-y",
	}
	stdout, _, err = counterpartyChain.Exec(ctx, cmd, nil)
	require.NoError(t, err)
	require.NotEmpty(t, stdout)
	testutil.WaitForBlocks(ctx, 1, counterpartyChain)

	cmd = []string{
		counterpartyChain.Config().Bin, "q", "gov", "vote", textProp.ProposalID, counterpartyChainUser.FormattedAddress(),
		"--node", counterpartyChain.GetRPCAddress(),
		"--home", counterpartyChain.HomeDir(),
		"--chain-id", counterpartyChain.Config().ChainID,
		"--output", "json",
	}
	stdout, _, err = counterpartyChain.Exec(ctx, cmd, nil)
	require.NoError(t, err, "could not query the vote")
	require.NotEmpty(t, stdout)

	var propResponse QueryVoteResponse
	err = json.Unmarshal(stdout, &propResponse)
	require.NoError(t, err)
	t.Log(propResponse.Options[0].Option)

	// SubmitTx on contract which will vote on the proposal
	execMsg = `{"vote":{"proposal_id":` + textProp.ProposalID + `,"option":1}}`
	t.Log(execMsg)
	cmd = []string{
		"archwayd", "tx", "wasm", "execute", contractAddress, execMsg,
		"--from", archwayChainUser.KeyName(), "--keyring-backend", keyring.BackendTest,
		"--gas", "auto", "--gas-prices", "0" + archwayChain.Config().Denom, "--gas-adjustment", "2",
		"--node", archwayChain.GetRPCAddress(),
		"--home", archwayChain.HomeDir(),
		"--chain-id", archwayChain.Config().ChainID,
		"--output", "json",
		"-y",
	}
	stdout, _, err = archwayChain.Exec(ctx, cmd, nil)
	require.NoError(t, err)
	require.NotEmpty(t, stdout)

	var txRes cosmos.CosmosTx
	err = json.Unmarshal(stdout, &txRes)
	require.NoError(t, err)
	t.Log(txRes.TxHash)

	err = testutil.WaitForBlocks(ctx, 1, archwayChain)
	require.NoError(t, err)

	cmd = []string{
		"archwayd", "q", "tx", txRes.TxHash,
		"--node", archwayChain.GetRPCAddress(),
		"--home", archwayChain.HomeDir(),
		"--chain-id", archwayChain.Config().ChainID,
		"--output", "json",
	}
	stdout, _, err = archwayChain.Exec(ctx, cmd, nil)
	require.NoError(t, err, "could not query the tx")

	aH, err := archwayChain.Height(ctx)
	require.NoError(t, err)

	cH, err := counterpartyChain.Height(ctx)
	require.NoError(t, err)

	ir := cosmos.DefaultEncoding().InterfaceRegistry
	recv, err := cosmos.PollForMessage(ctx, counterpartyChain, ir, cH, cH+10, func(found *channeltypes.MsgRecvPacket) bool {
		return found.Packet.Sequence == 1
	})
	require.NoError(t, err)
	t.Log(string(recv.Packet.Data))

	ackFound := func(found *channeltypes.MsgAcknowledgement) bool {
		return found.Packet.Sequence == 1 &&
			found.Packet.DestinationPort == "icahost"
	}

	ack, err := cosmos.PollForMessage(ctx, archwayChain, ir, aH, aH+10, ackFound)
	require.NoError(t, err)
	t.Log(string(ack.Acknowledgement))

	// Wait for a while to ensure the relayer picks up the packet
	err = testutil.WaitForBlocks(ctx, 10, archwayChain, counterpartyChain)
	require.NoError(t, err)

	err = archwayChain.QueryContract(ctx, contractAddress, QueryMsg{DumpState: &struct{}{}}, &contractRes)
	require.NoError(t, err)
	require.NotNil(t, contractRes.Data)
	// Ensure the contract is in the expected state
	t.Log(contractRes.Data.Voted)

	t.Log("--------------------------------------------------------------------")
	cmd = []string{
		counterpartyChain.Config().Bin, "q", "gov", "votes", textProp.ProposalID, // icaCounterpartyAddress,
		"--node", counterpartyChain.GetRPCAddress(),
		"--home", counterpartyChain.HomeDir(),
		"--chain-id", counterpartyChain.Config().ChainID,
		"--output", "json",
	}
	stdout, _, err = counterpartyChain.Exec(ctx, cmd, nil)
	require.NoError(t, err, "could not query the vote")
	require.NotEmpty(t, stdout)

	t.Log(string(stdout))

	// err = json.Unmarshal(stdout, &propResponse)
	// require.NoError(t, err)
	// t.Log(propResponse.Options[0].Option)

	// // Check the balance of the ica account on counterparty chain. Should be none
	// balance, err = counterpartyChain.GetBalance(ctx, icaCounterpartyAddress, counterpartyChain.Config().Denom)
	// require.NoError(t, err)
	// require.Equal(t, int64(0), balance.Int64())

	// // Wait for a few blocks on both chains so relayer picks up the packet
	// err = testutil.WaitForBlocks(ctx, 10, archwayChain, counterpartyChain)
	// require.NoError(t, err)

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

type InterchainAccountAccountQueryResponse struct {
	InterchainAccountAddress string `json:"interchain_account_address"`
}

type ContractResponse struct {
	Data ContractResponseObj `json:"data"`
}

type ContractResponseObj struct {
	Count               int32  `json:"count"`
	Owner               string `json:"owner"`
	ConnectionId        string `json:"connection_id"`
	CounterpartyVersion string `json:"counterparty_version"`
	Voted               bool   `json:"voted"`
}

type QueryMsg struct {
	DumpState *struct{} `json:"dump_state"`
}

type QueryVoteResponse struct {
	ProposalID string       `json:"proposal_id"`
	Voter      string       `json:"voter"`
	Options    []VoteOption `json:"options"`
	Metadata   string       `json:"metadata"`
}

type VoteOption struct {
	Option string `json:"option"`
	Weight string `json:"weight"`
}
