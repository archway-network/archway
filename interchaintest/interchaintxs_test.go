package interchaintest

import (
	"context"
	"encoding/json"
	"testing"

	"cosmossdk.io/math"
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

	gaiaChainSpec := &interchaintest.ChainSpec{
		Name:      "juno",
		ChainName: "juno",
		Version:   "v20.0.0",
		ChainConfig: ibc.ChainConfig{
			UsingNewGenesisCommand: true,
		},
	}
	numOfVals := 1
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
		path        = "ibc-upgrade-test-path"
		relayerName = "relayer"
	)
	relayerFactory := interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t),
		relayer.StartupFlags("-b", "100"),
	)
	client, network := interchaintest.DockerSetup(t)
	relayer := relayerFactory.Build(t, client, network)

	// Create the IBC network with the chains and relayer
	ic := interchaintest.NewInterchain().
		AddChain(archwayChain).
		AddChain(counterpartyChain).
		AddRelayer(relayer, "relayer").
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
		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})
	err = testutil.WaitForBlocks(ctx, 5, archwayChain, counterpartyChain)
	require.NoError(t, err)

	// Upload the contract to archway chain
	archwayChainUser := fundChainUser(t, ctx, archwayChain)
	counterpartyChainUser := fundChainUser(t, ctx, counterpartyChain)
	codeId, err := archwayChain.StoreContract(ctx, archwayChainUser.KeyName(), "artifacts/contract.wasm")
	require.NoError(t, err)
	require.NotEmpty(t, codeId)
	// Instantiate the contract
	contractAddress, err := archwayChain.InstantiateContract(ctx, archwayChainUser.KeyName(), codeId, "{}", false)
	require.NoError(t, err)
	require.NotEmpty(t, contractAddress)
	// Execute the contract to register an ica account
	res, err := archwayChain.ExecuteContract(ctx, archwayChainUser.KeyName(), contractAddress, "{}")
	require.NoError(t, err)
	require.NotEmpty(t, res)
	// Wait for a few blocks on both chains so relayer picks up the packet
	err = testutil.WaitForBlocks(ctx, 5, archwayChain, counterpartyChain)
	require.NoError(t, err)
	// Check if ica has been registered on the counterparty chain
	ownerAddress := contractAddress
	connections, err := relayer.GetConnections(ctx, eRep, archwayChain.Config().ChainID)
	require.NoError(t, err)
	require.Len(t, connections, 1)
	interchainAccountId := "interchain-account-0"
	cmd := []string{
		"archwayd", "q", "interchaintxs", "interchain-account", ownerAddress, connections[0].ID, interchainAccountId,
		"--node", archwayChain.GetRPCAddress(),
		"--home", archwayChain.HomeDir(),
		"--chain-id", archwayChain.Config().ChainID,
		"--output", "json",
	}
	stdout, _, err := archwayChain.Exec(ctx, cmd, nil)
	require.NoError(t, err, "could not query the interchain account")
	require.NotEmpty(t, stdout)

	queryRes := InterchainAccountAccountQueryResponse{}
	err = json.Unmarshal(stdout, &queryRes)
	require.NoError(t, err, "could not parse the interchain account query response")
	require.NotEmpty(t, queryRes.InterchainAccountAddress)
	icaCounterpartyAddress := queryRes.InterchainAccountAddress
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
	// SubmitTx on contract which will send money back to the initial account
	res, err = archwayChain.ExecuteContract(ctx, archwayChainUser.KeyName(), contractAddress, "{}")
	require.NoError(t, err)
	require.NotEmpty(t, res)
	// Wait for a while to ensure the relayer picks up the packet
	err = testutil.WaitForBlocks(ctx, 5, archwayChain, counterpartyChain)
	require.NoError(t, err)
	// Check the balance of the ica account on counterparty chain. Should be none
	balance, err = counterpartyChain.GetBalance(ctx, icaCounterpartyAddress, counterpartyChain.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, int64(0), balance.Int64())
}

type InterchainAccountAccountQueryResponse struct {
	InterchainAccountAddress string `json:"interchain_account_address"`
}
