package interchaintest

import (
	"fmt"

	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
)

const (
	initialVersion = "v3.0.0" // The last release of the chain. The one the mainnet is running on
	upgradeName    = "v4.0.0" // The next upgrade name. Should match the upgrade handler.
	chainName      = "archway"
)

const (
	coinType          = "118"
	chainType         = "cosmos"
	ibcTrustingPeriod = "112h"

	denom        = "aarch"
	bech32Prefix = "archway"
)

func GetArchwaySpec(dockerImageVersion string, numOfVals int) *interchaintest.ChainSpec {
	chainConfig := getDefaultChainConfig()
	chainConfig.Images = []ibc.DockerImage{{
		Repository: chainName,
		Version:    dockerImageVersion,
		UidGid:     "1025:1025",
	}}
	archwayChainSpec := &interchaintest.ChainSpec{
		Name:          chainName,
		ChainName:     "archway-1",
		Version:       dockerImageVersion,
		ChainConfig:   chainConfig,
		NumValidators: &numOfVals,
	}
	return archwayChainSpec
}

func getDefaultChainConfig() ibc.ChainConfig {
	return ibc.ChainConfig{
		Type:                   chainType,
		Name:                   chainName,
		ChainID:                "archway-local",
		Bin:                    "archwayd",
		Bech32Prefix:           bech32Prefix,
		Denom:                  denom,
		CoinType:               coinType,
		GasPrices:              fmt.Sprintf("0%s", denom),
		GasAdjustment:          2.0,
		TrustingPeriod:         ibcTrustingPeriod,
		NoHostMount:            false,
		SkipGenTx:              false,
		PreGenesis:             nil,
		ModifyGenesis:          cosmos.ModifyGenesis(getTestGenesis()),
		ConfigFileOverrides:    nil,
		UsingNewGenesisCommand: false,
	}
}

const (
	votingPeriod     = "10s" // Reducing voting period for testing
	maxDepositPeriod = "10s" // Reducing max deposit period for testing
)

func getTestGenesis() []cosmos.GenesisKV {
	return []cosmos.GenesisKV{
		{
			Key:   "app_state.gov.voting_params.voting_period",
			Value: votingPeriod,
		},
		{
			Key:   "app_state.gov.deposit_params.max_deposit_period",
			Value: maxDepositPeriod,
		},
		{
			Key:   "app_state.gov.deposit_params.min_deposit.0.denom",
			Value: denom,
		},
		{
			Key:   "app_state.mint.params.mint_denom",
			Value: denom,
		},
		{
			Key:   "app_state.rewards.params.min_price_of_gas.denom",
			Value: denom,
		},
		{
			Key:   "app_state.rewards.min_consensus_fee.denom",
			Value: denom,
		},
		{
			Key:   "app_state.staking.params.bond_denom",
			Value: denom,
		},
	}
}
