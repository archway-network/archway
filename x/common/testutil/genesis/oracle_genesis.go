package genesis

// import (
// 	"cosmossdk.io/math"

// 	"github.com/archway-network/archway/app"
// 	"github.com/archway-network/archway/x/common/asset"
// 	"github.com/archway-network/archway/x/common/denoms"
// 	oracletypes "github.com/archway-network/archway/x/oracle/types"
// )

// func AddOracleGenesis(gen app.GenesisState) app.GenesisState {
// 	gen[oracletypes.ModuleName] = app.MakeEncodingConfig().Codec.
// 		MustMarshalJSON(OracleGenesis())
// 	return gen
// }

// func OracleGenesis() *oracletypes.GenesisState {
// 	oracleGenesis := oracletypes.DefaultGenesisState()
// 	oracleGenesis.ExchangeRates = []oracletypes.ExchangeRateTuple{
// 		{Pair: asset.Registry.Pair(denoms.ETH, denoms.NUSD), ExchangeRate: math.LegacyNewDec(1_000)},
// 		{Pair: asset.Registry.Pair(denoms.NIBI, denoms.NUSD), ExchangeRate: math.LegacyNewDec(10)},
// 	}
// 	oracleGenesis.Params.VotePeriod = 1_000

// 	return oracleGenesis
// }
