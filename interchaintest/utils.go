package interchaintest

import (
	"context"
	"encoding/json"
	"errors"

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
	// archwayd query interchain-accounts controller interchain-account cosmos1layxcsmyye0dc0har9sdfzwckaz8sjwlfsj8zs connection-0
	cmd := []string{
		chain.Config().Bin, "q", "interchain-accounts", "controller", "interchain-account", ownerAddress, connectionId,
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
	return queryRes.Address, err
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

func RegisterContractForError(chain *cosmos.CosmosChain, user ibc.Wallet, ctx context.Context, contractAddress string) error {
	getSubscriptionFeesCmd := []string{
		chain.Config().Bin, "q",
		"cwerrors", "params",
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--output", "json",
	}
	stdout, _, err := chain.Exec(ctx, getSubscriptionFeesCmd, nil)
	if err != nil {
		return err
	}

	var params CwErrorParams
	if err = json.Unmarshal(stdout, &params); err != nil {
		return err
	}

	errorRegistrationCmd := []string{
		chain.Config().Bin, "tx",
		"cwerrors", "subscribe", contractAddress, params.SubscriptionFee.String(),
		"--from", user.KeyName(), "--keyring-backend", keyring.BackendTest,
		"--gas", "auto", "--gas-prices", "0aarch", "--gas-adjustment", "2",
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--output", "json",
		"-y",
	}
	_, _, err = chain.Exec(ctx, errorRegistrationCmd, nil)
	if err != nil {
		return err
	}
	checkSubscriptionCmd := []string{
		chain.Config().Bin, "q",
		"cwerrors", "is-subscribed", contractAddress,
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--output", "json",
	}
	stdout, _, err = chain.Exec(ctx, checkSubscriptionCmd, nil)
	if err != nil {
		return err
	}
	var isSubscribedResp CwErrorIsSubscribed
	if err = json.Unmarshal(stdout, &isSubscribedResp); err != nil {
		return err
	}
	if isSubscribedResp.Subscribed {
		return nil
	}
	return errors.New("contract is not subscribed")
}

func GetStoredCWErrors(chain *cosmos.CosmosChain, ctx context.Context, contractAddress string) ([]CWError, error) {
	cmd := []string{
		chain.Config().Bin, "q", "cwerrors", "errors", contractAddress,
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--output", "json",
	}
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	if err != nil {
		return nil, err
	}
	var errorsResp CWErrorResponse
	err = json.Unmarshal(stdout, &errorsResp)
	return errorsResp.Errors, err
}
