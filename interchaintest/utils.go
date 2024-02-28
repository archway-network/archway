package interchaintest

import (
	"context"
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"gopkg.in/yaml.v2"
)

func InstantiateContract(chain *cosmos.CosmosChain, user ibc.Wallet, ctx context.Context, codeId string, initMsg string) (string, error) {
	// Instantiate the contract
	cmd := []string{
		chain.Config().Bin, "tx", "wasm", "instantiate", codeId, initMsg,
		"--label", "cwica-contract", "--admin", user.FormattedAddress(),
		"--from", user.KeyName(), "--keyring-backend", keyring.BackendTest,
		"--gas", "auto", "--gas-prices", "0aarch", "--gas-adjustment", "2",
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--output", "json",
		"-y",
	}
	if _, _, err := chain.Exec(ctx, cmd, nil); err != nil {
		return "", err
	}
	if err := testutil.WaitForBlocks(ctx, 1, chain); err != nil {
		return "", err
	}

	// Getting the contract address
	cmd = []string{
		chain.Config().Bin, "q", "wasm", "list-contract-by-code", codeId,
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
	}
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	if err != nil {
		return "", err
	}
	contactsRes := cosmos.QueryContractResponse{}
	if err = yaml.Unmarshal(stdout, &contactsRes); err != nil {
		return "", err
	}
	return contactsRes.Contracts[0], nil
}

func ExecuteContract(chain *cosmos.CosmosChain, user ibc.Wallet, ctx context.Context, contractAddress string, execMsg string) error {
	cmd := []string{
		chain.Config().Bin, "tx", "wasm", "execute", contractAddress, execMsg,
		"--from", user.KeyName(), "--keyring-backend", keyring.BackendTest,
		"--gas", "auto", "--gas-prices", "0aarch", "--gas-adjustment", "2",
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--output", "json",
		"-y",
	}
	_, _, err := chain.Exec(ctx, cmd, nil)
	return err
}

func GetInterchainAccountAddress(chain *cosmos.CosmosChain, ctx context.Context, ownerAddress string, connectionId string) (string, error) {
	cmd := []string{
		chain.Config().Bin, "q", "cwica", "interchain-account", ownerAddress, connectionId,
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--output", "json",
	}
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	if err != nil {
		return "", err
	}
	var queryRes InterchainAccountAccountQueryResponse
	err = json.Unmarshal(stdout, &queryRes)
	return queryRes.InterchainAccountAddress, err
}

func GetUserVote(chain *cosmos.CosmosChain, ctx context.Context, proposalId string, address string) (QueryVoteResponse, error) {
	cmd := []string{
		chain.Config().Bin, "q", "gov", "vote", proposalId, address,
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--output", "json",
	}
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	if err != nil {
		return QueryVoteResponse{}, err
	}
	var propResponse QueryVoteResponse
	err = json.Unmarshal(stdout, &propResponse)
	return propResponse, err
}
