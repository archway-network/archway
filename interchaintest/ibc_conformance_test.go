package interchaintest

import (
	"context"
	"testing"

	interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/conformance"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/relayer"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestGaiaConformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	gaiaChainSpec := &interchaintest.ChainSpec{
		Name:      "gaia",
		ChainName: "gaia",
		Version:   "latest",
	}
	runConformanceTest(t, gaiaChainSpec)
}

func runConformanceTest(t *testing.T, gaiaChainSpec *interchaintest.ChainSpec) {
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
	ctx := context.Background()

	// Starts all the components of the IBC network
	require.NoError(t, ic.Build(ctx, rep.RelayerExecReporter(t), interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	// Perform the conformance test between the two chains
	conformance.TestChainPair(t, ctx, client, network, archwayChain, counterpartyChain, relayerFactory, rep, relayer, path)
}
