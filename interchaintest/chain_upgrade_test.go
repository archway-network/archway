package interchaintest

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	cosmosproto "github.com/cosmos/gogoproto/proto"
	"github.com/docker/docker/client"
	interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const (
	haltHeightDelta    = uint64(10) // The number of blocks after which to apply upgrade after creation of proposal.
	blocksAfterUpgrade = uint64(10) // The number of blocks to wait for after the upgrade has been applied.
)

func TestChainUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	archwayChain, client, ctx := startChain(t, initialVersion)
	chainUser := fundChainUser(t, ctx, archwayChain)
	haltHeight := submitUpgradeProposalAndVote(t, ctx, upgradeName, archwayChain, chainUser)

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

func submitUpgradeProposalAndVote(t *testing.T, ctx context.Context, nextUpgradeName string, archwayChain *cosmos.CosmosChain, chainUser ibc.Wallet) uint64 {
	height, err := archwayChain.Height(ctx) // The current chain height
	require.NoError(t, err, "error fetching height before submit upgrade proposal")

	haltHeight := height + haltHeightDelta // The height at which upgrade should be applied

	govAuthorityAddr := ""
	cmd := []string{
		"archwayd", "q", "auth", "module-account", "gov",
		"--node", archwayChain.GetRPCAddress(),
		"--home", archwayChain.HomeDir(),
		"--chain-id", archwayChain.Config().ChainID,
		"--output", "json",
	}
	stdout, _, err := archwayChain.Exec(ctx, cmd, nil)
	require.NoError(t, err, "could not query the gov module account")

	queryRes := ModuleAccountQueryResponse{}
	err = json.Unmarshal(stdout, &queryRes)
	require.NoError(t, err, "could not parse the response")

	if queryRes.Account.Name == "gov" {
		govAuthorityAddr = queryRes.Account.BaseAccount.Address
	} else {
		t.Fatal("could not find the gov module account")
	}

	proposalMsg := upgradetypes.MsgSoftwareUpgrade{
		Authority: govAuthorityAddr,
		Plan: upgradetypes.Plan{
			Name:   nextUpgradeName,
			Height: int64(haltHeight),
		},
	}

	proposal, err := archwayChain.BuildProposal([]cosmosproto.Message{&proposalMsg},
		"Test Upgrade",
		"Every PR we preform an upgrade check to ensure nothing breaks",
		"metadata",
		"10000000000"+archwayChain.Config().Denom,
	)
	require.NoError(t, err, "error building proposal tx")

	upgradeTx, err := archwayChain.SubmitProposal(ctx, chainUser.KeyName(), proposal) // Submitting the software upgrade proposal
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

func startChain(t *testing.T, startingVersion string) (*cosmos.CosmosChain, *client.Client, context.Context) {
	numOfVals := 1
	archwayChainSpec := GetArchwaySpec(initialVersion, numOfVals)
	archwayChainSpec.ChainConfig.ModifyGenesis = cosmos.ModifyGenesis(getTestGenesis())
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		archwayChainSpec,
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

type ModuleAccountQueryResponse struct {
	Account ModuleAccountData `json:"account"`
}

type ModuleAccountData struct {
	BaseAccount BaseAccountData `json:"base_account"`
	Name        string          `json:"name"`
}

type BaseAccountData struct {
	AccountNumber string `json:"account_number"`
	Address       string `json:"address"`
	PubKey        string `json:"pub_key"`
	Sequence      string `json:"sequence"`
}
