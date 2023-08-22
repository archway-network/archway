package interchaintest

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/client"
	interchaintest "github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

const (
	haltHeightDelta    = uint64(10) // The number of blocks after which to apply upgrade after creation of proposal.
	blocksAfterUpgrade = uint64(10) // The number of blocks to wait for after the upgrade has been applied.
)

func TestChainUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	archwayChain, client, ctx := startChain(t)
	chainUser := fundChainUser(t, ctx, archwayChain)
	haltHeight := submitUpgradeProposalAndVote(t, ctx, archwayChain, chainUser)

	height, err := archwayChain.Height(ctx)
	require.NoError(t, err, "cound not fetch height before upgrade")

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*45)
	defer timeoutCtxCancel()

	// This should timeout due to chain halt at upgrade height.
	_ = testutil.WaitForBlocks(timeoutCtx, int(haltHeight-height)+1, archwayChain)

	height, err = archwayChain.Height(ctx)
	require.NoError(t, err, "could not fetch height after chain should have halted")

	// Make sure that chain is halted
	require.Equal(t, haltHeight, height, "height is not equal to halt height")

	// Bring down nodes to prepare for upgrade
	err = archwayChain.StopAllNodes(ctx)
	require.NoError(t, err, "could not stop node(s)")

	// Upgrade version on all nodes - We are passing in the local image for the upgrade build
	archwayChain.UpgradeVersion(ctx, client, chainName, "local")

	// Start all nodes back up.
	// Validators reach consensus on first block after upgrade height
	// And chain block production resumes 🎉
	err = archwayChain.StartAllNodes(ctx)
	require.NoError(t, err, "could not start upgraded node(s)")

	timeoutCtx, timeoutCtxCancel = context.WithTimeout(ctx, time.Second*45)
	defer timeoutCtxCancel()

	err = testutil.WaitForBlocks(timeoutCtx, int(blocksAfterUpgrade), archwayChain)
	require.NoError(t, err, "chain did not produce blocks after upgrade")
}

func submitUpgradeProposalAndVote(t *testing.T, ctx context.Context, archwayChain *cosmos.CosmosChain, chainUser ibc.Wallet) uint64 {
	height, err := archwayChain.Height(ctx) // The current chain height
	require.NoError(t, err, "error fetching height before submit upgrade proposal")

	haltHeight := height + haltHeightDelta // The height at which upgrade should be applied

	proposal := cosmos.SoftwareUpgradeProposal{
		Deposit:     "10000000000" + archwayChain.Config().Denom,
		Title:       "Test upgrade",
		Name:        upgradeName,
		Description: "Every PR we perform a upgrade check to ensure nothing breaks",
		Height:      haltHeight,
	}

	upgradeTx, err := archwayChain.UpgradeProposal(ctx, chainUser.KeyName(), proposal) // Submitting the software upgrade proposal
	require.NoError(t, err, "error submitting software upgrade proposal tx")

	err = archwayChain.VoteOnProposalAllValidators(ctx, upgradeTx.ProposalID, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	_, err = cosmos.PollForProposalStatus(ctx, archwayChain, height, height+haltHeightDelta, upgradeTx.ProposalID, cosmos.ProposalStatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")
	return haltHeight
}

func fundChainUser(t *testing.T, ctx context.Context, archwayChain *cosmos.CosmosChain) ibc.Wallet {
	const userFunds = int64(10_000_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, archwayChain)
	return users[0]
}

func startChain(t *testing.T) (*cosmos.CosmosChain, *client.Client, context.Context) {
	numOfVals := 1
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          chainName,
			ChainName:     "archway-1",
			Version:       initialVersion,
			ChainConfig:   archwayConfig,
			NumValidators: &numOfVals,
		},
	})
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)
	archwayChain := chains[0].(*cosmos.CosmosChain)

	ic := interchaintest.NewInterchain().AddChain(archwayChain)
	client, network := interchaintest.DockerSetup(t)
	ctx := context.Background()
	require.NoError(t, ic.Build(ctx, nil, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: true,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})
	return archwayChain, client, ctx
}
