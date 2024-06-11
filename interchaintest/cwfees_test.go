package interchaintest

import (
	"testing"

	"cosmossdk.io/math"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/stretchr/testify/require"
)

// TestCWFees tests the CWFees functionality.
func TestCWFees(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	archwayChain, _, ctx := startChain(t, initialVersion)
	cwgranterUser := interchaintest.GetAndFundTestUsers(t, ctx, "granter", math.NewInt(10_000_000_000_000), archwayChain)[0]
	cwgranteeUser := interchaintest.GetAndFundTestUsers(t, ctx, "cwgrantee", math.NewInt(1), archwayChain)[0]

	// Upload the cwfees granter contract to archway chain
	codeID, err := archwayChain.StoreContract(ctx, cwgranterUser.KeyName(), "artifacts/cwfees.wasm")
	require.NoError(t, err)
	require.NotEmpty(t, codeID)

	// Instantiate the contract
	initMsg := "{\"allowed_address\":\"" + cwgranteeUser.FormattedAddress() + "\"}"
	contractAddress, err := InstantiateContract(archwayChain, cwgranterUser, ctx, codeID, initMsg)
	require.NoError(t, err)

	// Send a msg with a cw fee granter
	// The cwgranteeUser is sending the one token they have to the cwgranterUser, using their address as the fee payer
	fromAddress := cwgranteeUser.FormattedAddress()
	toAddress := cwgranterUser.FormattedAddress()
	cmd := []string{
		archwayChain.Config().Bin, "tx", "bank", "send", fromAddress, toAddress, "1aarch",
		"--node", archwayChain.GetRPCAddress(),
		"--home", archwayChain.HomeDir(),
		"--fee-granter", contractAddress,
		"--chain-id", archwayChain.Config().ChainID,
	}
	stdout, _, err := archwayChain.Exec(ctx, cmd, nil)
	require.NoError(t, err, "failed to send tx")
	t.Log(stdout)

	// Ensure the balances match
	// cwgrantee should have zero balance
	// cwgranter should have balance - gasfees + 1 token

	// Ensure tx fails when the sender is not whitelisted/the contract refuses to pay for this users fees

	// Ensure malicious contract which uses more gas than allowed does not break anything

	// Confirm cosmos-sdk/x/feegrant behaviour
	// ensure that our custom feegrant logic didnt mess up the sdk feegrant logic
}
