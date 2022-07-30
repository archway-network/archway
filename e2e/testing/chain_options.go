package e2eTesting

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/archway-network/archway/app"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

// chainConfig is a TestChain config which can be adjusted using options.
type chainConfig struct {
	ValidatorsNum    int
	GenAccountsNum   int
	GenBalanceAmount string
	BondAmount       string
}

type (
	TestChainConfigOption func(cfg *chainConfig)

	TestChainGenesisOption func(cdc codec.Codec, genesis app.GenesisState)

	TestChainConsensusParamsOption func(params *abci.ConsensusParams)
)

// WithBlockGasLimit sets the block gas limit (not set by default).
func WithBlockGasLimit(gasLimit int64) TestChainConsensusParamsOption {
	return func(params *abci.ConsensusParams) {
		params.Block.MaxGas = gasLimit
	}
}

// WithInflationRewardsRatio sets x/rewards inflation rewards ratio parameter.
func WithInflationRewardsRatio(ratio sdk.Dec) TestChainGenesisOption {
	return func(cdc codec.Codec, genesis app.GenesisState) {
		var rewardsGenesis rewardsTypes.GenesisState
		cdc.MustUnmarshalJSON(genesis[rewardsTypes.ModuleName], &rewardsGenesis)

		rewardsGenesis.Params.InflationRewardsRatio = ratio

		genesis[rewardsTypes.ModuleName] = cdc.MustMarshalJSON(&rewardsGenesis)
	}
}

// WithTxFeeRebatesRewardsRatio sets x/rewards tx fee rebates rewards ratio parameter.
func WithTxFeeRebatesRewardsRatio(ratio sdk.Dec) TestChainGenesisOption {
	return func(cdc codec.Codec, genesis app.GenesisState) {
		var rewardsGenesis rewardsTypes.GenesisState
		cdc.MustUnmarshalJSON(genesis[rewardsTypes.ModuleName], &rewardsGenesis)

		rewardsGenesis.Params.TxFeeRebateRatio = ratio

		genesis[rewardsTypes.ModuleName] = cdc.MustMarshalJSON(&rewardsGenesis)
	}
}
