package interchaintest

import (
	"context"
	"testing"
	"time"

	"gopkg.in/yaml.v2"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/testutil"
	"github.com/stretchr/testify/require"
)

// TestAccountBurnChainUpgrade was written specifically to test for the following issue : https://github.com/orgs/archway-network/discussions/6
// To run this test, you will need the following heighliner images
// heighliner build --org archway-network --repo archway --dockerfile cosmos --build-target "make build" --build-env "BUILD_TAGS=muslc" --binaries "build/archwayd" --git-ref v2.0.0 --tag v2.0.0 -c archway
// heighliner build --org archway-network --repo archway --dockerfile cosmos --build-target "make build" --build-env "BUILD_TAGS=muslc" --binaries "build/archwayd" --git-ref v4.0.0 --tag v4.0.0 -c archway
// heighliner build --org archway-network --repo archway --dockerfile cosmos --build-target "make build" --build-env "BUILD_TAGS=muslc" --binaries "build/archwayd" --git-ref v4.0.1 --tag v4.0.1 -c archway
func TestFeeCollectorBurnChainUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	// Starting the chain with v2.0.0. Starting at v2.0.0 because bug only happens when we have upgraded to v4.0.0. Does not happen when we start from there.
	archwayChain, client, ctx := startChain(t, "v2.0.0")
	chainUser := fundChainUser(t, ctx, archwayChain)

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*45)
	defer timeoutCtxCancel()

	// waiting for chain starting
	testutil.WaitForBlocks(timeoutCtx, 1, archwayChain)

	// Ensuring feecollector does not have burn permissions in v2.0.0
	queryRes2 := getModuleAccount(t, ctx, authtypes.FeeCollectorName, archwayChain)
	require.Len(t, queryRes2.Account.Permissions, 0, "feecollector should not have burn permissions in v2.0.0")

	// Upgrading to v4.0.0 => Not directly upgrading to v4.0.1 to simulate how things went on constantine.
	haltHeight := submitUpgradeProposalAndVote(t, ctx, "v4.0.0", archwayChain, chainUser)
	height, err := archwayChain.Height(ctx)
	require.NoError(t, err, "cound not fetch height before upgrade")
	testutil.WaitForBlocks(timeoutCtx, int(haltHeight-height)+1, archwayChain)
	height, err = archwayChain.Height(ctx)
	require.NoError(t, err, "could not fetch height after chain should have halted")
	require.Equal(t, int(haltHeight), int(height), "height is not equal to halt height")
	err = archwayChain.StopAllNodes(ctx)
	require.NoError(t, err, "could not stop node(s)")
	archwayChain.UpgradeVersion(ctx, client, chainName, "v4.0.0")
	err = archwayChain.StartAllNodes(ctx)
	require.NoError(t, err, "could not start upgraded node(s)")
	timeoutCtx, timeoutCtxCancel = context.WithTimeout(ctx, time.Second*45)
	defer timeoutCtxCancel()
	err = testutil.WaitForBlocks(timeoutCtx, int(blocksAfterUpgrade), archwayChain)
	require.NoError(t, err, "chain did not produce blocks after upgrade")

	// Ensuring feecollector does not have burn permissions in v4.0.0
	queryRes4 := getModuleAccount(t, ctx, authtypes.FeeCollectorName, archwayChain)
	require.Len(t, queryRes4.Account.Permissions, 0, "feecollector should not have burn permissions in v4.0.0")

	// Upgrading to v4.0.1
	haltHeight = submitUpgradeProposalAndVote(t, ctx, "v4.0.1", archwayChain, chainUser)
	height, err = archwayChain.Height(ctx)
	require.NoError(t, err, "cound not fetch height before upgrade")
	testutil.WaitForBlocks(timeoutCtx, int(haltHeight-height)+1, archwayChain)
	height, err = archwayChain.Height(ctx)
	require.NoError(t, err, "could not fetch height after chain should have halted")
	require.Equal(t, int(haltHeight), int(height), "height is not equal to halt height")
	err = archwayChain.StopAllNodes(ctx)
	require.NoError(t, err, "could not stop node(s)")
	archwayChain.UpgradeVersion(ctx, client, chainName, "v4.0.1")
	err = archwayChain.StartAllNodes(ctx)
	require.NoError(t, err, "could not start upgraded node(s)")
	timeoutCtx, timeoutCtxCancel = context.WithTimeout(ctx, time.Second*45)
	defer timeoutCtxCancel()
	err = testutil.WaitForBlocks(timeoutCtx, int(blocksAfterUpgrade), archwayChain)
	require.NoError(t, err, "chain did not produce blocks after upgrade")

	// Ensuring feecollector HAS burn permissions in v4.0.1
	queryRes401 := getModuleAccount(t, ctx, authtypes.FeeCollectorName, archwayChain)
	require.Len(t, queryRes401.Account.Permissions, 1, "feecollector should have one permissions in v4.0.1")
	require.Equal(t, authtypes.Burner, queryRes401.Account.Permissions[0], "feecollector should have burn permissions in v4.0.1")
}

func getModuleAccount(t *testing.T, ctx context.Context, moduleAccountName string, archwayChain *cosmos.CosmosChain) QueryModuleAccountResponse {
	cmd := []string{
		"archwayd", "q", "auth", "module-account", moduleAccountName,
		"--node", archwayChain.GetRPCAddress(),
		"--home", archwayChain.HomeDir(),
		"--chain-id", archwayChain.Config().ChainID,
	}
	stdout, _, err := archwayChain.Exec(ctx, cmd, nil)
	require.NoError(t, err, "could not fetch the fee collector account")
	queryRes := QueryModuleAccountResponse{}
	err = yaml.Unmarshal(stdout, &queryRes)
	require.NoError(t, err, "could not unmarshal query module account respons")
	return queryRes
}

type QueryModuleAccountResponse struct {
	Account AccountData `json:"account"`
}

type AccountData struct {
	Name        string   `json:"name`
	Permissions []string `json:"permissions`
}
