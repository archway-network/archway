package interchaintest

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
)

const (
	initialVersion = "v6.0.3" // The last release of the chain. The one the mainnet is running on
	upgradeName    = "latest" // The next upgrade name. Should match the upgrade handler.
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
		Type:           chainType,
		Name:           chainName,
		ChainID:        "archway-local",
		Bin:            "archwayd",
		Bech32Prefix:   bech32Prefix,
		Denom:          denom,
		CoinType:       coinType,
		GasPrices:      fmt.Sprintf("0%s", denom),
		GasAdjustment:  2.0,
		TrustingPeriod: ibcTrustingPeriod,
		NoHostMount:    false,
		SkipGenTx:      false,
		PreGenesis:     nil,
		ModifyGenesisAmounts: func() (types.Coin, types.Coin) {
			genesisAmount := types.Coin{
				Amount: types.NewInt(9_000_000_000_000_000_000),
				Denom:  denom,
			}
			genesisSelfDelegation := types.Coin{
				Amount: types.NewInt(5_000_000_000_000_000_000),
				Denom:  denom,
			}
			return genesisAmount, genesisSelfDelegation
		},
	}
}

const (
	votingPeriod     = "10s" // Reducing voting period for testing
	maxDepositPeriod = "10s" // Reducing max deposit period for testing
)

func getTestGenesis() []cosmos.GenesisKV {
	return []cosmos.GenesisKV{
		{
			Key:   "app_state.gov.params.voting_period",
			Value: votingPeriod,
		},
		{
			Key:   "app_state.gov.params.max_deposit_period",
			Value: maxDepositPeriod,
		},
	}
}
