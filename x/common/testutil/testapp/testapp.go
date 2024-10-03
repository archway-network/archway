package testapp

import (
	"encoding/json"

	"time"

	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	tmdb "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/archway-network/archway/app"
	"github.com/archway-network/archway/app/appconst"
	// epochstypes "github.com/archway-network/archway/x/epochs/types"
	// inflationtypes "github.com/archway-network/archway/x/inflation/types"
	// sudotypes "github.com/archway-network/archway/x/sudo/types"
)

// NewNibiruTestAppAndContext creates an 'app.ArchwayApp' instance with an
// in-memory 'tmdb.MemDB' and fresh 'sdk.Context'.
func NewNibiruTestAppAndContext() (*app.ArchwayApp, sdk.Context) {
	// Set up base app
	encoding := app.MakeEncodingConfig()
	var appGenesis app.GenesisState = app.NewDefaultGenesisState(encoding.Marshaler)
	// genModEpochs := epochstypes.DefaultGenesisFromTime(time.Now().UTC())

	// // Set happy genesis: epochs
	// appGenesis[epochstypes.ModuleName] = encoding.Codec.MustMarshalJSON(
	// 	genModEpochs,
	// )

	// Set happy genesis: sudo
	// sudoGenesis := new(sudotypes.GenesisState)
	// sudoGenesis.Sudoers = DefaultSudoers()
	// appGenesis[sudotypes.ModuleName] = encoding.Codec.MustMarshalJSON(sudoGenesis)

	app := NewNibiruTestApp(appGenesis)
	ctx := NewContext(app)

	// Set defaults for certain modules.
	// app.OracleKeeper.SetPrice(ctx, asset.Registry.Pair(denoms.BTC, denoms.NUSD), math.LegacyNewDec(20000))
	// app.OracleKeeper.SetPrice(ctx, "xxx:yyy", math.LegacyNewDec(20000))
	// app.SudoKeeper.Sudoers.Set(ctx, DefaultSudoers())

	return app, ctx
}

// NewContext: Returns a fresh sdk.Context corresponding to the given NibiruApp.
func NewContext(nibiru *app.ArchwayApp) sdk.Context {
	return nibiru.NewContext(false)
}

// DefaultSudoers: State for the x/sudo module for the default test app.
// func DefaultSudoers() sudotypes.Sudoers {
// 	addr := DefaultSudoRoot().String()
// 	return sudotypes.Sudoers{
// 		Root:      addr,
// 		Contracts: []string{addr},
// 	}
// }

// SetDefaultSudoGenesis: Sets the sudo module genesis state to a valid
// default. See "DefaultSudoers".
func SetDefaultSudoGenesis(gen app.GenesisState) {
	// sudoGen := new(sudotypes.GenesisState)
	// encoding := app.MakeEncodingConfig()
	// encoding.Codec.MustUnmarshalJSON(gen[sudotypes.ModuleName], sudoGen)
	// if err := sudoGen.Validate(); err != nil {
	// 	sudoGen.Sudoers = DefaultSudoers()
	// 	gen[sudotypes.ModuleName] = encoding.Codec.MustMarshalJSON(sudoGen)
	// }
}

// NewNibiruTestAppAndZeroTimeCtx: Runs NewNibiruTestAppAndZeroTimeCtx with the
// block time set to time zero.
func NewNibiruTestAppAndContextAtTime(startTime time.Time) (*app.ArchwayApp, sdk.Context) {
	app, _ := NewNibiruTestAppAndContext()
	ctx := NewContext(app).WithBlockTime(startTime)
	return app, ctx
}

// NewNibiruTestApp initializes a chain with the given genesis state to
// creates an application instance ('app.ArchwayApp'). This app uses an
// in-memory database ('tmdb.MemDB') and has logging disabled.
func NewNibiruTestApp(gen app.GenesisState, baseAppOptions ...func(*baseapp.BaseApp)) *app.ArchwayApp {
	db := tmdb.NewMemDB()
	logger := log.NewNopLogger()

	encoding := app.MakeEncodingConfig()
	SetDefaultSudoGenesis(gen)

	application := app.NewArchwayApp(
		logger,
		db,
		/*traceStore=*/ nil,
		/*loadLatest=*/ true,
		/*skipUpgradeHeights=*/ map[int64]bool{},
		/*homePath=*/ appconst.DefaultNodeHome,
		/*invCheckPeriod=*/ 0,
		encoding,
		/*appOpts=*/ sims.EmptyAppOptions{},
		/*wasmOpts=*/ nil,
		baseAppOptions...,
	)

	gen, err := GenesisStateWithSingleValidator(encoding.Marshaler, gen)
	if err != nil {
		panic(err)
	}

	stateBytes, err := json.MarshalIndent(gen, "", " ")
	if err != nil {
		panic(err)
	}

	application.InitChain(&abci.RequestInitChain{
		ConsensusParams: sims.DefaultConsensusParams,
		AppStateBytes:   stateBytes,
	})

	return application
}

// FundAccount is a utility function that funds an account by minting and
// sending the coins to the address. This should be used for testing purposes
// only!
func FundAccount(
	bankKeeper bankkeeper.Keeper, ctx sdk.Context, addr sdk.AccAddress,
	amounts sdk.Coins,
) error {
	if err := bankKeeper.MintCoins(ctx, ibctransfertypes.ModuleName, amounts); err != nil {
		return err
	}

	return bankKeeper.SendCoinsFromModuleToAccount(ctx, ibctransfertypes.ModuleName, addr, amounts)
}

// FundModuleAccount is a utility function that funds a module account by
// minting and sending the coins to the address. This should be used for testing
// purposes only!
func FundModuleAccount(
	bankKeeper bankkeeper.Keeper, ctx sdk.Context,
	recipientMod string, amounts sdk.Coins,
) error {
	// if err := bankKeeper.MintCoins(ctx, inflationtypes.ModuleName, amounts); err != nil {
	// 	return err
	// }

	// return bankKeeper.SendCoinsFromModuleToModule(ctx, inflationtypes.ModuleName, recipientMod, amounts)
	return nil
}
