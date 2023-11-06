package e2eTesting

import (
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	archway "github.com/archway-network/archway/types"

	"github.com/archway-network/archway/app"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

// chainConfig is a TestChain config which can be adjusted using options.
type chainConfig struct {
	ValidatorsNum    int
	GenAccountsNum   int
	GenBalanceAmount string
	BondAmount       string
	LoggerEnabled    bool
	DefaultFeeAmt    string
	DummyTestAddr    bool
}

type (
	TestChainConfigOption func(cfg *chainConfig)

	TestChainGenesisOption func(cdc codec.Codec, genesis app.GenesisState)

	TestChainConsensusParamsOption func(params *tmproto.ConsensusParams)
)

// defaultChainConfig builds chain default config.
func defaultChainConfig() chainConfig {
	return chainConfig{
		ValidatorsNum:    1,
		GenAccountsNum:   5,
		GenBalanceAmount: archway.DefaultPowerReduction.MulRaw(100).String(),
		BondAmount:       archway.DefaultPowerReduction.MulRaw(1).String(),
		DefaultFeeAmt:    archway.DefaultPowerReduction.QuoRaw(10).String(), // 0.1
		DummyTestAddr:    false,
	}
}

func WithValidatorsNum(num int) TestChainConfigOption {
	return func(cfg *chainConfig) {
		cfg.ValidatorsNum = num
	}
}

func WithDummyTestAddress() TestChainConfigOption {
	return func(cfg *chainConfig) {
		cfg.DummyTestAddr = true
	}
}

// WithGenAccounts sets the number of genesis accounts
func WithGenAccounts(num int) TestChainConfigOption {
	return func(cfg *chainConfig) {
		cfg.GenAccountsNum = num
	}
}

// WithGenDefaultCoinBalance sets the genesis account balance for the default token (stake).
func WithGenDefaultCoinBalance(amount string) TestChainConfigOption {
	return func(cfg *chainConfig) {
		cfg.GenBalanceAmount = amount
	}
}

// WithDefaultFeeAmount sets the default fee amount which is used for sending Msgs with no fee specified.
func WithDefaultFeeAmount(amount string) TestChainConfigOption {
	return func(cfg *chainConfig) {
		cfg.DefaultFeeAmt = amount
	}
}

// WithBondAmount sets the amount of coins to bond for each validator.
func WithBondAmount(amount string) TestChainConfigOption {
	return func(cfg *chainConfig) {
		cfg.BondAmount = amount
	}
}

// WithLogger enables the app console logger.
func WithLogger() TestChainConfigOption {
	return func(cfg *chainConfig) {
		cfg.LoggerEnabled = true
	}
}

// WithBlockGasLimit sets the block gas limit (not set by default).
func WithBlockGasLimit(gasLimit int64) TestChainConsensusParamsOption {
	return func(params *tmproto.ConsensusParams) {
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

// WithMaxWithdrawRecords sets x/rewards MaxWithdrawRecords param.
func WithMaxWithdrawRecords(num uint64) TestChainGenesisOption {
	return func(cdc codec.Codec, genesis app.GenesisState) {
		var rewardsGenesis rewardsTypes.GenesisState
		cdc.MustUnmarshalJSON(genesis[rewardsTypes.ModuleName], &rewardsGenesis)

		rewardsGenesis.Params.MaxWithdrawRecords = num

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

// WithMintParams sets x/mint inflation calculation parameters.
func WithMintParams(inflationMin, inflationMax sdk.Dec, blocksPerYear uint64) TestChainGenesisOption {
	return func(cdc codec.Codec, genesis app.GenesisState) {
		var mintGenesis mintTypes.GenesisState
		cdc.MustUnmarshalJSON(genesis[mintTypes.ModuleName], &mintGenesis)

		mintGenesis.Params.InflationMin = inflationMin
		mintGenesis.Params.InflationMax = inflationMax
		mintGenesis.Params.BlocksPerYear = blocksPerYear

		genesis[mintTypes.ModuleName] = cdc.MustMarshalJSON(&mintGenesis)
	}
}
