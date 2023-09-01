package interchaintest

import (
	"fmt"

	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
)

const (
	initialVersion = "v4.0.2" // The last release of the chain. The one the mainnet is running on
	upgradeName    = "latest" // The next upgrade name. Should match the upgrade handler.
	chainName      = "archway"
)

const (
	votingPeriod     = "10s" // Reducing voting period for testing
	maxDepositPeriod = "10s" // Reducing max deposit period for testing
)

var (
	coinType = "118"
	denom    = "aarch"

	dockerImage = ibc.DockerImage{
		Repository: chainName,
		Version:    initialVersion,
		UidGid:     "1025:1025",
	}

	archwayConfig = ibc.ChainConfig{
		Type:                   "cosmos",
		Name:                   chainName,
		ChainID:                "archway-local",
		Images:                 []ibc.DockerImage{dockerImage},
		Bin:                    "archwayd",
		Bech32Prefix:           "archway",
		Denom:                  denom,
		CoinType:               coinType,
		GasPrices:              fmt.Sprintf("0%s", denom),
		GasAdjustment:          2.0,
		TrustingPeriod:         "112h",
		NoHostMount:            false,
		SkipGenTx:              false,
		PreGenesis:             nil,
		ModifyGenesis:          cosmos.ModifyGenesis(getTestGenesis()),
		ConfigFileOverrides:    nil,
		UsingNewGenesisCommand: false,
	}
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
